package app

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	clientv1 "github.com/joelanford/workshop/pkg/client/v1"
	"github.com/joelanford/workshop/pkg/workshopcontroller"
)

var AppVersion string

func Run(logger log.Logger) error {
	app := cli.NewApp()

	app.Version = AppVersion

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "kubeconfig, c",
			Value:  filepath.Join(os.Getenv("HOME"), ".kube", "config"),
			EnvVar: "KUBECONFIG",
			Usage:  "load kubernetes config from `FILE`",
		},
		cli.BoolFlag{
			Name:  "in-cluster, i",
			Usage: "set if running in a kubernetes cluster",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "clean",
			Aliases: []string{"c"},
			Usage:   "clean up workshop resources and exit",
			Action: func(c *cli.Context) error {
				// load the client config.
				client, err := getClient(c)
				if err != nil {
					return errors.Wrapf(err, "could not load workshop client")
				}

				// create the controller
				controller := workshopcontroller.New(logger, client)

				// Clean up the workshop resources
				if err := controller.Clean(); err != nil {
					logger.Log("msg", "could not delete custom resource definition", "err", err)
				}
				logger.Log("msg", "successfully deleted custom resource definition")
				return nil
			},
		},
	}

	app.Before = func(c *cli.Context) error {
		inCluster := c.IsSet("in-cluster")
		if inCluster {
			logger.Log("in-cluster", inCluster)
		} else {
			filepathAbs, err := filepath.Abs(c.String("kubeconfig"))
			if err != nil {
				return errors.Wrap(err, "could not get path of kubeconfig")
			}
			logger.Log("kubeconfig", filepathAbs)
		}
		return nil
	}

	app.Action = func(c *cli.Context) error {
		// load the client config.
		client, err := getClient(c)
		if err != nil {
			return errors.Wrapf(err, "could not load workshop client")
		}

		// create the controller
		controller := workshopcontroller.New(logger, client)

		ctx, cancel := context.WithCancel(context.Background())
		wg, ctx := errgroup.WithContext(ctx)

		// run the controller
		wg.Go(func() error { return controller.Run(ctx) })

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

		// wait until the controller is done or we get a SIGTERM
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGTERM:
				logger.Log("msg", "received SIGTERM, exiting gracefully...")
			case syscall.SIGUSR1:
				logger.Log("msg", "received SIGUSR1, cleaning up and exiting gracefully...")
				if err := controller.Clean(); err != nil {
					logger.Log("msg", "could not delete custom resource definition", "err", err)
				}
				logger.Log("msg", "successfully deleted custom resource definition")
			}
		case <-ctx.Done():
		}

		cancel()
		if err := wg.Wait(); err != nil && err != context.Canceled {
			return errors.Wrapf(err, "unhandled error received")
		}
		return nil
	}

	return app.Run(os.Args)
}

func getClient(c *cli.Context) (*clientv1.Client, error) {
	if c.IsSet("in-cluster") {
		return clientv1.NewFromCluster()
	} else {
		return clientv1.NewFromFile(c.String("kubeconfig"))
	}
}
