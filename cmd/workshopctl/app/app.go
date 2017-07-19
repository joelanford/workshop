package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/urfave/cli"

	workshopv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	"github.com/joelanford/workshop/pkg/workshopctl"
)

var (
	version   string
	buildTime string
	buildUser string
	gitHash   string
)

func Run() error {
	cli.VersionPrinter = printVersion
	app := cli.NewApp()

	app.Name = "workshopctl"
	app.HelpName = "workshopctl"
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
	}
	app.Commands = []cli.Command{
		{
			Name:  "create",
			Usage: "create a new workshop resource",
			Subcommands: cli.Commands{
				{
					Name:    "desk",
					Aliases: []string{"d"},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "version, v",
							Value: workshopv1.DefaultDeskVersion,
							Usage: "create bench with version `VERSION`",
						},
						cli.StringFlag{
							Name:  "lifespan, l",
							Value: workshopv1.DefaultDeskLifespan.String(),
							Usage: "duration of desk lifespan",
						},
					},
					Action: workshopctl.CreateDesk(),
				},
			},
		},
		{
			Name:  "get",
			Usage: "get workshop resources",
			Subcommands: cli.Commands{
				{
					Name:    "desk",
					Aliases: []string{"desks", "d"},
					Action:  workshopctl.GetDesk(),
				},
			},
		},
		{
			Name:  "delete",
			Usage: "delete a workshop resource",
			Subcommands: cli.Commands{
				{
					Name:    "desk",
					Aliases: []string{"desks", "d"},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "all",
							Usage: "delete all desks`",
						},
					},
					Action: workshopctl.DeleteDesk(),
				},
			}},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	return app.Run(os.Args)
}

func printVersion(c *cli.Context) {
	fmt.Printf("Version:     %s\nBuild Time:  %s\nBuild User:  %s\nGit Hash:    %s\n", version, buildTime, buildUser, gitHash)
}
