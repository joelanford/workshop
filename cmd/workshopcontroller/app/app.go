package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	clientv1 "github.com/joelanford/workshop/pkg/client/v1"
	"github.com/joelanford/workshop/pkg/workshopcontroller"
)

var (
	version   string
	buildTime string
	buildUser string
	gitHash   string
)

func Run(logger log.Logger) error {
	cli.VersionPrinter = printVersion
	app := cli.NewApp()

	app.Name = "workshopcontroller"
	app.HelpName = "workshopcontroller"
	app.Version = version

	if compiled, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", buildTime); err == nil {
		app.Compiled = compiled
	}

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
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "setup custom resource definitions and run the controller",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "expiration-interval, e",
					Usage: "interval (in seconds) of expiration checks",
					Value: 60,
				},
			},
			Action: func(c *cli.Context) error {
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
				expirationInterval := time.Second * time.Duration(c.Int("expiration-interval"))
				wg.Go(func() error { return controller.Run(ctx, expirationInterval) })

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
			},
		},
	}

	return app.Run(os.Args)
}

func getClient(c *cli.Context) (*clientv1.Client, error) {
	if c.IsSet("in-cluster") {
		return clientv1.NewFromCluster()
	}
	return clientv1.NewFromFile(c.String("kubeconfig"))
}

func printVersion(c *cli.Context) {
	fmt.Printf("Version:     %s\nBuild Time:  %s\nBuild User:  %s\nGit Hash:    %s\n", version, buildTime, buildUser, gitHash)
}
