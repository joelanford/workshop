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
		var (
			client *clientv1.Client
			err    error
		)
		if c.IsSet("in-cluster") {
			client, err = clientv1.NewFromCluster()
		} else {
			client, err = clientv1.NewFromFile(c.String("kubeconfig"))
		}
		if err != nil {
			return errors.Wrapf(err, "could not load create workshop client")
		}

		// create the controller
		controller := workshopcontroller.New(logger, client)

		ctx, cancel := context.WithCancel(context.Background())
		wg, ctx := errgroup.WithContext(ctx)

		// run the controller
		wg.Go(func() error { return controller.Run(ctx) })

		term := make(chan os.Signal)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)

		// wait until the controller is done or we get a SIGTERM
		select {
		case <-term:
			logger.Log("msg", "received SIGTERM, exiting gracefully...")
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
