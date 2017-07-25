package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/joelanford/workshop/cmd/workshop-controller/app/glogshim"
	"github.com/joelanford/workshop/pkg/workshop/controller"
)

var (
	version   string
	buildTime string
	buildUser string
	gitHash   string
)

func Run() error {
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
		cli.DurationFlag{
			Name:  "initial-sync-timeout, t",
			Value: time.Minute,
			Usage: "timeout for initial resource sync.",
		},
		cli.StringFlag{
			Name:  "domain, d",
			Usage: "if set, will use as domain suffix for workshop services.",
		},
	}
	app.Flags = append(app.Flags, glogshim.Flags...)

	app.Action = func(c *cli.Context) error {
		glogshim.ShimCLI(c)

		domain := c.String("domain")
		kubeconfig := c.String("kubeconfig")
		initialSyncTimeout := c.Duration("initial-sync-timeout")
		healthzPort := c.Int("healthz-port")
		clean := c.IsSet("clean")

		wc, err := controller.NewWorkshopController(kubeconfig, domain, initialSyncTimeout)
		if err != nil {
			return err
		}

		if clean {
			glog.V(0).Infof("Cleaning workshop resources and exiting.")
			return wc.Clean()
		}

		if domain != "" {
			if ok := validateDomain(domain); !ok {
				return fmt.Errorf("Invalid domain: %s", domain)
			}
		}

		for _, flagName := range c.GlobalFlagNames() {
			glog.V(0).Infof("FLAG --%s=%q", flagName, c.Generic(flagName))
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
			glog.V(0).Infof("Starting healthz server at %v with readiness handler at \"/readiness\"", healthzPort)
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
			glog.V(0).Infof("Could not shutdown healthz server gracefully: %s. Forcing shutdown", err)
			healthzServer.Close()
		}
		shutdownCancel()

		if err := wg.Wait(); err != nil && err != context.Canceled && err != http.ErrServerClosed {
			return errors.Wrap(err, "Unhandled error received")
		}
		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	return app.Run(os.Args)
}

func printVersion(c *cli.Context) {
	fmt.Printf("Version:     %s\nBuild Time:  %s\nBuild User:  %s\nGit Hash:    %s\n", version, buildTime, buildUser, gitHash)
}

func validateDomain(domain string) bool {
	parsed, err := url.Parse(domain)
	if err != nil {
		return false
	}
	if parsed.Scheme != "" || parsed.Path != "" || strings.Contains(parsed.Host, ":") {
		return false
	}
	return true
}
