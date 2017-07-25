package controller

import (
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/pkg/api/v1"
	kcache "k8s.io/client-go/tools/cache"

	"fmt"

	"github.com/golang/glog"
	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

func (c *WorkshopController) setDesksStore() {
	// Returns a cache.ListWatch that gets all changes to desks.
	c.desksStore, c.desksController = kcache.NewInformer(
		kcache.NewListWatchFromClient(
			c.workshopClient.WorkshopV1().RESTClient(),
			"desks",
			v1.NamespaceAll,
			fields.Everything()),
		&apiv1.Desk{},
		resyncPeriod,
		kcache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleDeskAdd,
			UpdateFunc: c.handleDeskUpdate,
			DeleteFunc: c.handleDeskDelete,
		},
	)
}

func (c *WorkshopController) waitForDesksSynced() error {
	// Wait for controllers to have completed an initial resource listing
	timeout := time.After(c.initialSyncTimeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timeout waiting for initialization")
		case <-ticker.C:
			if c.desksController.HasSynced() {
				glog.V(0).Infof("Initialized desks from apiserver")
				return nil
			}
			glog.V(0).Infof("Waiting for desks to be initialized from apiserver...")
		}
	}
}

func (c *WorkshopController) handleDeskAdd(obj interface{}) {
	if d, ok := obj.(*apiv1.Desk); ok {
		if err := c.createDeskResources(d); err != nil {
			glog.Errorf("Error in createDesk(%v): %v", d.Name, err)
		}
	}
}

func (c *WorkshopController) handleDeskUpdate(oldObj, newObj interface{}) {
	if d, ok := newObj.(*apiv1.Desk); ok {
		if err := c.updateDeskResources(d); err != nil {
			glog.Errorf("Error in updateDesk(%v): %v", d.Name, err)
		}
	}
}

func (c *WorkshopController) handleDeskDelete(obj interface{}) {
	if d, ok := obj.(*apiv1.Desk); ok {
		if err := c.deleteDeskResources(d); err != nil {
			glog.Errorf("Error in deleteDesk(%v): %v", d.Name, err)
		}
	}
}

func (c *WorkshopController) createDeskResources(desk *apiv1.Desk) error {
	glog.V(0).Infof("Creating resources for desk \"%s\"", desk.Name)

	trustedNamespace, err := c.createDeskNamespace(desk, fmt.Sprintf("%s-desk-trusted", desk.Name))
	if err != nil {
		return err
	}

	defaultNamespace, err := c.createDeskNamespace(desk, fmt.Sprintf("%s-desk-default", desk.Name))
	if err != nil {
		return err
	}

	sa, err := c.createDeskServiceAccount(desk, trustedNamespace)
	if err != nil {
		return err
	}

	_, err = c.createDeskRoleBinding(desk, "view", sa, trustedNamespace)
	if err != nil {
		return err
	}

	_, err = c.createDeskRoleBinding(desk, "edit", sa, defaultNamespace)
	if err != nil {
		return err
	}

	_, err = c.createDeskKubeshellDeployment(desk, trustedNamespace, defaultNamespace)
	if err != nil {
		return err
	}

	_, err = c.createDeskKubeshellService(desk, trustedNamespace)
	if err != nil {
		return err
	}

	if c.domain != "" {
		_, err = c.createDeskKubeshellIngress(desk, trustedNamespace, c.domain)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *WorkshopController) updateDeskResources(desk *apiv1.Desk) error {
	glog.V(0).Infof("Updating resources for desk \"%s\"", desk.Name)
	return nil
}

func (c *WorkshopController) deleteDeskResources(desk *apiv1.Desk) error {
	glog.V(0).Infof("Deleting resources for desk \"%s\"", desk.Name)

	if err := c.deleteDeskNamespace(desk, fmt.Sprintf("%s-desk-trusted", desk.Name)); err != nil {
		return err
	}

	if err := c.deleteDeskNamespace(desk, fmt.Sprintf("%s-desk-default", desk.Name)); err != nil {
		return err
	}

	return nil
}
