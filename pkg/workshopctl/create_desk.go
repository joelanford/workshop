package workshopctl

import (
	"errors"
	"fmt"
	"time"

	"github.com/urfave/cli"

	workshopv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	clientv1 "github.com/joelanford/workshop/pkg/client/v1"
)

func CreateDesk() cli.ActionFunc {
	return func(c *cli.Context) error {
		client, err := clientv1.NewFromFile(c.GlobalString("kubeconfig"))
		if err != nil {
			return err
		}

		if c.NArg() == 0 {
			return errors.New("NAME is required")
		}
		name := c.Args()[0]

		owner := name
		version := c.String("version")
		if version == "" {
			version = workshopv1.DeskDefaultVersion
		}

		lifespanStr := c.String("lifespan")
		lifespan, err := time.ParseDuration(lifespanStr)
		if err != nil {
			lifespan = workshopv1.DeskMaxLifespan
		}
		expiration := time.Now().Add(lifespan)

		desk, err := client.CreateDesk(name, owner, version, expiration)
		if err != nil {
			return err
		}
		fmt.Printf("desk \"%s\" created\n", desk.ObjectMeta.Name)
		return nil
	}
}
