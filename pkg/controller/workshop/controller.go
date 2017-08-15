package workshop

import (
	"context"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	kcache "k8s.io/client-go/tools/cache"

	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	"github.com/joelanford/workshop/pkg/client/workshop"
	"github.com/joelanford/workshop/pkg/controller/desk"
)

type Controller struct {
	domain             string
	initialSyncTimeout time.Duration

	kubeClient     kubernetes.Interface
	apiExtClient   apiextensionsclient.Interface
	workshopClient workshop.Interface

	deskCache resourceCache

	logger *logrus.Logger
}

type resourceCache struct {
	store      kcache.Store
	controller kcache.Controller
}

func NewController(config *rest.Config, domain string, logger *logrus.Logger) (*Controller, error) {
	c := &Controller{
		domain: domain,
		logger: logger,
	}
	if err := c.setClients(config); err != nil {
		return nil, err
	}
	c.setDeskCache()
	return c, nil
}

func (c *Controller) Start(ctx context.Context) error {
	c.logger.Info("creating desk custom resource definition")
	if err := c.createDeskCRD(); err != nil {
		if apierrors.IsAlreadyExists(err) {
			c.logger.Info("desk custom resource definition already exists; continuing execution")
		} else {
			c.logger.Fatalf("could not create desk custom resource definition: %s", err)
		}
	}

	c.cleanStaleResources()

	c.logger.Infof("Starting desksController")
	go c.deskCache.controller.Run(ctx.Done())

	return c.waitForDesksSynced()
}

func (c *Controller) setClients(config *rest.Config) error {
	var err error
	c.kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.apiExtClient, err = apiextensionsclient.NewForConfig(config)
	if err != nil {
		return err
	}

	c.workshopClient, err = workshop.NewForConfig(config)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) setDeskCache() {
	c.deskCache.store, c.deskCache.controller = kcache.NewInformer(
		kcache.NewListWatchFromClient(
			c.workshopClient.WorkshopV1().RESTClient(),
			"desks",
			v1.NamespaceAll,
			fields.Everything()),
		&apiv1.Desk{},
		5*time.Minute,
		desk.NewEventHandler(c.kubeClient, c.workshopClient, c.logger),
	)
}

func (c *Controller) createDeskCRD() error {
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiv1.DeskCRDName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   apiv1.GroupName,
			Version: apiv1.SchemeGroupVersion.Version,
			Scope:   apiextensionsv1beta1.ClusterScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural: apiv1.DeskResourcePlural,
				Kind:   apiv1.DeskKind,
			},
		},
	}
	if _, err := c.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
		return err
	}
	return nil
}

func (c *Controller) deleteDeskCRD() error {
	return c.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(apiv1.DeskCRDName, nil)
}

func (c *Controller) cleanStaleResources() {
	c.logger.Infof("Cleaning up desk resources whose desk is no longer present...")
	nsList, err := c.kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		c.logger.Errorf("Error listing namespaces for staleness check and removal: %s", err)
	}

	deskList, err := c.workshopClient.WorkshopV1().Desks().List(metav1.ListOptions{})
	if err != nil {
		c.logger.Errorf("Error listing desks for staleness check and removal: %s", err)
	}
	deskMap := make(map[string]apiv1.Desk)
	for _, desk := range deskList.Items {
		deskMap[desk.Name] = desk
	}

	for _, namespace := range nsList.Items {
	ownerRefsLoop:
		for _, ownerRef := range namespace.OwnerReferences {
			if ownerRef.Kind == apiv1.DeskKind {
				if _, ok := deskMap[ownerRef.Name]; !ok {
					if err := c.kubeClient.CoreV1().Namespaces().Delete(namespace.Name, nil); err != nil {
						c.logger.Errorf("Error removing stale namespace \"%s\": %s", namespace.Name, err)
						break ownerRefsLoop
					}
					c.logger.Infof("Removed stale namespace \"%s\", which was owned by non-existent desk \"%s\"", namespace.Name, ownerRef.Name)
				}
			}
		}
	}
}

func (c *Controller) waitForDesksSynced() error {
	// Wait for controllers to have completed an initial resource listing
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for initialization")
		case <-ticker.C:
			if c.deskCache.controller.HasSynced() {
				c.logger.Infof("initialized desks from apiserver")
				return nil
			}
			c.logger.Infof("waiting for desks to be initialized from apiserver...")
		}
	}
}
