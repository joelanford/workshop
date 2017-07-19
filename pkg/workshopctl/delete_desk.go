package workshopctl

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"

	clientv1 "github.com/joelanford/workshop/pkg/client/v1"
)

func DeleteDesk() cli.ActionFunc {
	return func(c *cli.Context) error {
		client, err := clientv1.NewFromFile(c.GlobalString("kubeconfig"))
		if err != nil {
			return err
		}

		if c.IsSet("all") {
			deskList, err := client.ListDesks()
			if err != nil {
				return err
			}
			for _, desk := range deskList.Items {
				name := desk.GetName()

				if err := client.DeleteDesk(name); err != nil {
					return err
				}

				fmt.Printf("desk \"%s\" deleted\n", name)
			}
		} else if c.NArg() > 0 {
			name := c.Args()[0]

			if err := client.DeleteDesk(name); err != nil {
				return err
			}

			fmt.Printf("desk \"%s\" deleted\n", name)
			return nil
		} else {
			return errors.New("NAME or -all option is required")
		}
		return nil
	}
}
