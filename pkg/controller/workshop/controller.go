package workshop

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/Sirupsen/logrus"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/joelanford/workshop/pkg/client/workshop"
	"github.com/joelanford/workshop/pkg/controller/desk"
)

type Controller struct {
	deskController *desk.Controller
}

func NewController(config *rest.Config, domain string, logger *logrus.Logger) (*Controller, error) {
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	apiExtClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	workshopClient, err := workshop.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	c := &Controller{
		deskController: desk.NewController(kubeClient, apiExtClient, workshopClient, domain, logger),
	}

	return c, nil
}

func (c *Controller) Start(ctx context.Context) error {
	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return c.deskController.Start(ctx)
	})
	return wg.Wait()
}

func (c *Controller) Clean() error {
	return c.deskController.Clean()
}
