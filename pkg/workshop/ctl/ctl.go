package ctl

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	"github.com/joelanford/workshop/pkg/client/workshop"
)

type WorkshopctlCommand struct {
	workshopClient workshop.Interface
}

func NewWorkshopctlCommand() *WorkshopctlCommand {
	return &WorkshopctlCommand{}
}

func (c *WorkshopctlCommand) Initialize(kubeconfig string) error {
	var (
		config *rest.Config
		err    error
	)
	defaultKubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return err
		}
	} else if _, err := os.Stat(defaultKubeconfig); err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
		if err != nil {
			return err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			config, _ = clientcmd.BuildConfigFromFlags("http://localhost:8080", "")
		}
	}

	c.workshopClient, err = workshop.NewForConfig(config)
	return err
}

func (c *WorkshopctlCommand) CreateDesk(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("NAME is required")
	}
	name := ctx.Args()[0]

	owner := name
	version := ctx.String("version")
	if version == "" {
		version = apiv1.DeskDefaultVersion
	}

	expirationDurationStr := ctx.String("expiration")
	expirationDuration, err := time.ParseDuration(expirationDurationStr)
	if err != nil {
		expirationDuration = apiv1.DeskMaxLifespan
	}
	expiration := time.Now().Add(expirationDuration)

	desk, err := c.workshopClient.WorkshopV1().Desks().Create(&apiv1.Desk{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiv1.DeskSpec{
			Owner:               owner,
			Version:             version,
			ExpirationTimestamp: metav1.NewTime(expiration),
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("desk \"%s\" created\n", desk.ObjectMeta.Name)
	return nil
}

func (c *WorkshopctlCommand) GetDesk(ctx *cli.Context) error {
	var desks []apiv1.Desk
	if ctx.NArg() == 0 {
		deskList, err := c.workshopClient.WorkshopV1().Desks().List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		desks = deskList.Items
	} else {
		name := ctx.Args()[0]
		desk, err := c.workshopClient.WorkshopV1().Desks().Get(name, metav1.GetOptions{})
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
	fmt.Fprintln(&w, "NAME\tOWNER\tVERSION\tEXPIRATION")
	for _, desk := range desks {
		fmt.Fprintf(&w, "%s\t%s\t%s\t%s\n", desk.ObjectMeta.Name, desk.Spec.Owner, desk.Spec.Version, desk.Spec.ExpirationTimestamp)
	}
	return w.Flush()
}

func (c *WorkshopctlCommand) DeleteDesk(ctx *cli.Context) error {
	var names []string

	if ctx.IsSet("all") {
		deskList, err := c.workshopClient.WorkshopV1().Desks().List(metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, desk := range deskList.Items {
			name := desk.Name
			names = append(names, name)
		}
	} else if ctx.NArg() > 0 {
		names = ctx.Args()
	} else {
		return errors.New("NAME or --all option is required")
	}
	c.deleteDesksByName(names)
	return nil
}

func (c *WorkshopctlCommand) deleteDesksByName(names []string) {
	if len(names) == 0 {
		fmt.Println("No resources found.")
		return
	}

	for _, name := range names {
		if err := c.workshopClient.WorkshopV1().Desks().Delete(name, nil); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("desk \"%s\" deleted\n", name)
		}
	}
}
