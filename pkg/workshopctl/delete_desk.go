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

		var names []string

		if c.IsSet("all") {
			deskList, err := client.ListDesks()
			if err != nil {
				return err
			}

			for _, desk := range deskList.Items {
				name := desk.GetName()
				names = append(names, name)
			}
		} else if c.NArg() > 0 {
			names = c.Args()
		} else {
			return errors.New("NAME or -all option is required")
		}
		deleteDesksByName(client, names)
		return nil
	}
}

func deleteDesksByName(client *clientv1.Client, names []string) {
	if len(names) == 0 {
		fmt.Println("No resources found.")
		return
	}

	for _, name := range names {
		if err := client.DeleteDesk(name); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("desk \"%s\" deleted\n", name)
		}
	}
}
