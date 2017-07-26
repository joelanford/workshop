package controller

import (
	"context"
	"os"
	"path/filepath"
	"time"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/joelanford/workshop/pkg/client/workshop"
)

//
// TODO:
//   - Create a new struct type to store shared desk information.
//   - Automatically re-create deleted resources owned by a desk
//     (e.g. namespace, role binding, serviceaccount, deployment, etc.)
//

const (
	// Resync period for the kube controller loop.
	resyncPeriod = 5 * time.Minute
)

type WorkshopController struct {
	domain             string
	initialSyncTimeout time.Duration

	kubeClient     kubernetes.Interface
	apiExtClient   apiextensionsclient.Interface
	workshopClient workshop.Interface

	desksStore      kcache.Store
	desksController kcache.Controller
}

func NewWorkshopController(kubeconfig string, domain string, timeout time.Duration) (*WorkshopController, error) {
	c := &WorkshopController{
		domain:             domain,
		initialSyncTimeout: timeout,
	}
	if err := c.setClients(kubeconfig); err != nil {
		return nil, err
	}
	c.setDesksStore()
	return c, nil
}

func (c *WorkshopController) Start(ctx context.Context) error {
	glog.V(1).Infof("Creating desk custom resource definitions")
	if err := c.createDeskCRD(); err != nil {
		if apierrors.IsAlreadyExists(err) {
			glog.V(1).Infoln("Desk custom resource definition already exists, continuing execution")
		} else {
			glog.Fatalf("Could not create Desk custom resource definition: %v", err)
		}
	}
	glog.V(2).Infof("Starting desksController")
	go c.desksController.Run(ctx.Done())

	//
	// TODO:
	//       If a desk was deleted while the controller was not running, this
	//       controller won't know about it as is. To fix that, we need to
	//       check to see if any namespaces are owned by a desk that no longer
	//       exists. If we find any, we should delete them. Deleting a
	//       namespace will also remove all of the resources in that namespace,
	//       so we don't need to check other namespaced resource types.
	//

	// Wait synchronously for the initial list operations to be
	// complete of desks from APIServer.
	return c.waitForDesksSynced()
}

func (c *WorkshopController) Clean() error {
	glog.V(1).Infof("Deleting desk custom resource definition")
	return c.deleteDeskCRD()
}

func (c *WorkshopController) setClients(kubeconfig string) error {
	var (
		config *rest.Config
		err    error
	)

	defaultKubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else if _, err := os.Stat(defaultKubeconfig); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
	} else {
		if config, err = rest.InClusterConfig(); err != nil {
			config, _ = clientcmd.BuildConfigFromFlags("http://localhost:8080", "")
		}
	}

	c.kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.apiExtClient, err = apiextensionsclient.NewForConfig(config)
	if err != nil {
		return err
	}

	c.workshopClient, err = workshop.NewForConfig(config)
	if err != nil {
		return err
	}
	return nil
}
