package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"golang.org/x/sync/errgroup"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/joelanford/workshop/pkg/controller/workshop"
)

var (
	version   string
	buildTime string
	buildUser string
	gitHash   string
)

func Run(logger *logrus.Logger) error {
	cli.VersionPrinter = printVersion
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version",
		Usage: "print the version",
	}

	app := cli.NewApp()

	app.Name = "workshop-controller"
	app.HelpName = "workshop-controller"
	app.Version = version

	if compiled, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", buildTime); err == nil {
		app.Compiled = compiled
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "kubeconfig, k",
			EnvVar: "KUBECONFIG",
			Usage:  "load kubernetes config from `FILE`",
		},
		cli.BoolFlag{
			Name:  "clean, c",
			Usage: "clean workshop resources and exit",
		},
		cli.IntFlag{
			Name:  "healthz-port, p",
			Value: 8081,
			Usage: "port on which to serve a workshop-controller readiness probe.",
		},
		cli.StringFlag{
			Name:  "domain, d",
			Usage: "if set, will use as domain suffix for workshop services.",
		},
	}

	app.Action = func(c *cli.Context) error {
		domain := c.String("domain")
		kubeconfig := c.String("kubeconfig")
		healthzPort := c.Int("healthz-port")
		clean := c.IsSet("clean")

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

		wc, err := workshop.NewController(config, domain, logger)
		if err != nil {
			return err
		}

		if clean {
			logger.Infof("Cleaning workshop resources and exiting.")
			return wc.Clean()
		}

		for _, flagName := range c.GlobalFlagNames() {
			logger.Infof("FLAG --%s=%q", flagName, c.Generic(flagName))
		}

		ctx, cancel := context.WithCancel(context.Background())
		wg, ctx := errgroup.WithContext(ctx)

		mux := http.NewServeMux()
		mux.HandleFunc("/readiness", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "ok\n")
		})

		healthzServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", healthzPort),
			Handler: mux,
		}

		wg.Go(func() error {
			if err := wc.Start(ctx); err != nil {
				return err
			}
			logger.Infof("starting healthz server at %v with readiness handler at \"/readiness\"", healthzPort)
			return healthzServer.ListenAndServe()
		})

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigChan:
			logger.Infof("received signal %s, exiting gracefully", sig)
		case <-ctx.Done():
		}

		// Cancel the context used to run the workshop controller
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*5)
		if err := healthzServer.Shutdown(shutdownCtx); err != nil {
			logger.Infof("could not shutdown healthz server gracefully: %s, forcing shutdown", err)
			healthzServer.Close()
		}
		shutdownCancel()

		if err := wg.Wait(); err != nil && err != context.Canceled && err != http.ErrServerClosed {
			return errors.Wrap(err, "unhandled error received")
		}
		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	return app.Run(os.Args)
}

func printVersion(c *cli.Context) {
	fmt.Printf("Version:     %s\nBuild Time:  %s\nBuild User:  %s\nGit Hash:    %s\n", version, buildTime, buildUser, gitHash)
}
