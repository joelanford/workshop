package desk

import (
	"context"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/joelanford/workshop/pkg/client/workshop"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	kcache "k8s.io/client-go/tools/cache"

	workshopv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

type Controller struct {
	kubeClient     kubernetes.Interface
	workshopClient workshop.Interface

	crdManager *CRDManager

	desks           map[string]Manager
	cacheStore      kcache.Store
	cacheController kcache.Controller

	domain string
	logger *logrus.Logger
}

func NewController(kubeClient kubernetes.Interface, apiExtClient apiextensionsclient.Interface, workshopClient workshop.Interface, domain string, logger *logrus.Logger) *Controller {
	c := &Controller{
		kubeClient:     kubeClient,
		workshopClient: workshopClient,
		crdManager:     NewCRDManager(apiExtClient),
		desks:          make(map[string]Manager),
		domain:         domain,
		logger:         logger,
	}
	c.cacheStore, c.cacheController = kcache.NewInformer(
		kcache.NewListWatchFromClient(
			workshopClient.WorkshopV1().RESTClient(),
			"desks",
			v1.NamespaceAll,
			fields.Everything()),
		&workshopv1.Desk{},
		5*time.Minute,
		kcache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		},
	)
	return c
}

func (c *Controller) Start(ctx context.Context) error {
	// Create the Desk CRD and wait for it to be ready.
	c.logger.Info("creating desk custom resource definition...")
	if err := c.crdManager.CreateAndWait(); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		c.logger.Info("skipping: desk custom resource definition already exists")
	} else {
		c.logger.Info("created desk custom resource definition")
	}

	// Start the cache controller
	go c.cacheController.Run(ctx.Done())

	// Wait until the cache controller has synced.
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for !c.cacheController.HasSynced() {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for cache initialization")
		case <-ticker.C:
			c.logger.Info("waiting for desk cache to be initialized from apiserver...")
		}
	}
	c.logger.Info("successfully initialized desk cache")

	// Delete resources whose desk resource is no longer present.
	if err := c.deleteStaleResources(); err != nil {
		c.logger.Info("could not delete stale desk resources: %s", err)
	}
	return nil
}

func (c *Controller) Clean() error {
	return c.crdManager.Delete()
}

func (c *Controller) deleteStaleResources() error {
	nsList, err := c.kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing namespaces: %s", err)
	}

	deskList, err := c.workshopClient.WorkshopV1().Desks().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing desks: %s", err)
	}
	deskMap := make(map[string]workshopv1.Desk)
	for _, desk := range deskList.Items {
		deskMap[desk.Name] = desk
	}

	for _, namespace := range nsList.Items {
	ownerRefsLoop:
		for _, ownerRef := range namespace.OwnerReferences {
			if ownerRef.Kind == workshopv1.DeskKind {
				if _, ok := deskMap[ownerRef.Name]; !ok {
					if err := c.kubeClient.CoreV1().Namespaces().Delete(namespace.Name, nil); err != nil {
						c.logger.Errorf("could not remove stale namespace \"%s\": %s", namespace.Name, err)
						break ownerRefsLoop
					}
					c.logger.Infof("removed stale namespace \"%s\", which was owned by non-existent desk \"%s\"", namespace.Name, ownerRef.Name)
				}
			}
		}
	}
	return nil
}

func (c *Controller) onAdd(obj interface{}) {
	c.logger.Info("desk.Controller::onAdd()")
	if d, ok := obj.(*workshopv1.Desk); ok {
		desk := d.DeepCopyObject().(*workshopv1.Desk)
		deskManager := NewManager(c.kubeClient, c.domain, c.logger, desk)
		deskManager.Start()
		c.desks[desk.Name] = *deskManager
	}
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	c.logger.Info("desk.Controller::onUpdate()")
	if d, ok := newObj.(*workshopv1.Desk); ok {
		desk := d.DeepCopyObject().(*workshopv1.Desk)
		deskManager, ok := c.desks[desk.Name]
		if !ok {
			c.logger.Errorf("could not update desk resources for desk \"%s\": no controller found", d.Name)
			return
		}
		deskManager.Update(desk)
	}
}

func (c *Controller) onDelete(obj interface{}) {
	c.logger.Info("desk.Controller::onDelete()")
	if d, ok := obj.(*workshopv1.Desk); ok {
		deskManager, ok := c.desks[d.Name]
		if !ok {
			c.logger.Errorf("could not delete desk resources for desk \"%s\": no controller found", d.Name)
			return
		}
		deskManager.Stop()
		delete(c.desks, d.Name)
	}
}
