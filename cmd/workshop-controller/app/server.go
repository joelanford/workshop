package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/joelanford/workshop/cmd/workshop-controller/app/options"
	"github.com/joelanford/workshop/pkg/client/workshop"
	"github.com/joelanford/workshop/pkg/workshop/controller"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

type WorkshopControllerServer struct {
	healthzPort int
	clean       bool
	wc          *controller.WorkshopController
}

func NewWorkshopControllerServerDefault(config *options.WorkshopControllerConfig) *WorkshopControllerServer {
	kubeClient, apiExtClient, workshopClient, err := newClients(config)
	if err != nil {
		glog.Fatalf("Failed to create a kubernetes client: %v", err)
	}

	return &WorkshopControllerServer{
		healthzPort: config.HealthzPort,
		clean:       config.Clean,
		wc:          controller.NewWorkshopController(config.DesksDomain, kubeClient, apiExtClient, workshopClient, config.InitialSyncTimeout),
	}
}

func newClients(wcConfig *options.WorkshopControllerConfig) (kubernetes.Interface, apiextensionsclient.Interface, workshop.Interface, error) {
	var config *rest.Config
	var err error

	// If both kubeconfig and master URL are empty, use service account
	if wcConfig.KubeConfigFile == "" && wcConfig.KubeMasterURL == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags(
			wcConfig.KubeMasterURL, wcConfig.KubeConfigFile)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	apiExtClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	workshopClient, err := workshop.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}
	return kubeClient, apiExtClient, workshopClient, nil
}

func (server *WorkshopControllerServer) Run() {
	if server.clean {
		glog.V(0).Infof("Cleaning workshop resources and exiting.")

		if err := server.wc.Clean(); err != nil {
			glog.Fatal(err)
		}
		os.Exit(0)
	}

	pflag.VisitAll(func(flag *pflag.Flag) {
		glog.V(0).Infof("FLAG --%s=%q", flag.Name, flag.Value)
	})

	ctx, cancel := context.WithCancel(context.Background())
	wg, ctx := errgroup.WithContext(ctx)

	healthzServer := server.setupHealthzServer()

	wg.Go(func() error {
		if err := server.wc.Start(ctx); err != nil {
			return err
		}
		glog.V(0).Infof("Starting Healthz server at %v", server.healthzPort)
		return healthzServer.ListenAndServe()
	})

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		glog.V(0).Infof("Received signal %s, exiting gracefully", sig)
	case <-ctx.Done():
	}

	// Cancel the context used to run the workshop controller
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*5)
	if err := healthzServer.Shutdown(shutdownCtx); err != nil {
		glog.V(0).Infof("ould not shutdown healthz server gracefully: %s. Forcing shutdown", err)
		healthzServer.Close()
	}
	shutdownCancel()

	if err := wg.Wait(); err != nil && err != context.Canceled && err != http.ErrServerClosed {
		glog.Fatalf("Unhandled error received: %s", err)
	}
}

// setupHandlers sets up a readiness and liveness endpoint for workshop-controller.
func (server *WorkshopControllerServer) setupHealthzServer() *http.Server {
	glog.V(0).Infof("Setting up Healthz Handler (/readiness)")
	mux := http.NewServeMux()
	mux.HandleFunc("/readiness", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "ok\n")
	})

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", server.healthzPort),
		Handler: mux,
	}
}
