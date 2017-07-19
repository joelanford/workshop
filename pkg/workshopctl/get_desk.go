package workshopctl

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"text/tabwriter"

	workshopv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	clientv1 "github.com/joelanford/workshop/pkg/client/v1"
)

func GetDesk() cli.ActionFunc {
	return func(c *cli.Context) error {
		client, err := clientv1.NewFromFile(c.GlobalString("kubeconfig"))
		if err != nil {
			return err
		}

		var desks []workshopv1.Desk
		if c.NArg() == 0 {
			deskList, err := client.ListDesks()
			if err != nil {
				return err
			}
			desks = deskList.Items
		} else {
			name := c.Args()[0]
			desk, err := client.GetDesk(name)
			if err != nil {
				return err
			}
			desks = append(desks, *desk)
		}

		if len(desks) == 0 {
			fmt.Fprintf(os.Stdout, "No resources found.\n")
			return nil
		}

		var w tabwriter.Writer
		w.Init(os.Stdout, 0, 4, 6, ' ', 0)
		fmt.Fprintln(&w, "NAME\tSTATUS\tOWNER\tVERSION\tEXPIRATION")
		for _, desk := range desks {
			fmt.Fprintf(&w, "%s\t%s\t%s\t%s\t%s\n", desk.ObjectMeta.Name, desk.Status.State, desk.Spec.Owner, desk.Spec.Version, desk.Spec.ExpirationTimestamp)
		}
		return w.Flush()
	}
}
