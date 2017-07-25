package controller

import (
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		c.createDeskResources(d)
	}
}

func (c *WorkshopController) handleDeskUpdate(oldObj, newObj interface{}) {
	oldDesk, oldDeskOk := oldObj.(*apiv1.Desk)
	newDesk, newDeskOk := newObj.(*apiv1.Desk)
	if oldDeskOk && newDeskOk {
		c.updateDeskResources(oldDesk, newDesk)
	}
}

func (c *WorkshopController) handleDeskDelete(obj interface{}) {
	if d, ok := obj.(*apiv1.Desk); ok {
		c.deleteDeskResources(d)
	}
}

func (c *WorkshopController) createDeskResources(desk *apiv1.Desk) {
	glog.V(0).Infof("Creating resources for desk \"%s\"", desk.Name)

	trustedNamespaceName := fmt.Sprintf("%s-desk-trusted", desk.Name)
	trustedNamespace, err := c.createDeskNamespace(desk, trustedNamespaceName)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			glog.V(2).Info("Namespace \"%s\" for desk \"%s\" already exists", trustedNamespaceName, desk.Name)
			trustedNamespace, err = c.kubeClient.CoreV1().Namespaces().Get(trustedNamespaceName, metav1.GetOptions{})
		} else {
			glog.Error(err)
			return
		}
	}

	defaultNamespaceName := fmt.Sprintf("%s-desk-default", desk.Name)
	defaultNamespace, err := c.createDeskNamespace(desk, defaultNamespaceName)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			glog.V(2).Info("Namespace \"%s\" for desk \"%s\" already exists", defaultNamespaceName, desk.Name)
			defaultNamespace, err = c.kubeClient.CoreV1().Namespaces().Get(defaultNamespaceName, metav1.GetOptions{})
		} else {
			glog.Error(err)
			return
		}
	}

	saName := desk.Spec.Owner
	sa, err := c.createDeskServiceAccount(desk, saName, trustedNamespace)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			glog.V(2).Info("ServiceAccount \"%s\" for desk \"%s\" already exists", saName, desk.Name)
			sa, err = c.kubeClient.CoreV1().ServiceAccounts(trustedNamespaceName).Get(saName, metav1.GetOptions{})
		} else {
			glog.Error(err)
			return
		}
	}

	viewRbName := fmt.Sprintf("%s-%s", sa.Name, "view")
	_, err = c.createDeskRoleBinding(desk, viewRbName, "view", sa, trustedNamespace)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			glog.V(2).Info("RoleBinding \"%s\" for desk \"%s\" already exists", viewRbName, desk.Name)
		} else {
			glog.Error(err)
		}
	}

	editRbName := fmt.Sprintf("%s-%s", sa.Name, "edit")
	_, err = c.createDeskRoleBinding(desk, editRbName, "edit", sa, defaultNamespace)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			glog.V(2).Info("RoleBinding \"%s\" for desk \"%s\" already exists", editRbName, desk.Name)
		} else {
			glog.Error(err)
		}
	}

	kubeshellName := "kubeshell"
	_, err = c.createDeskKubeshellDeployment(desk, kubeshellName, trustedNamespace, defaultNamespace)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			glog.V(2).Info("Deployment \"%s\" for desk \"%s\" already exists", kubeshellName, desk.Name)
		} else {
			glog.Error(err)
		}
	}

	_, err = c.createDeskKubeshellService(desk, kubeshellName, trustedNamespace)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			glog.V(2).Info("Service \"%s\" for desk \"%s\" already exists", kubeshellName, desk.Name)
		} else {
			glog.Error(err)
		}
	}

	if c.domain != "" {
		_, err = c.createDeskKubeshellIngress(desk, kubeshellName, trustedNamespace, c.domain)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				glog.V(2).Info("Ingress \"%s\" for desk \"%s\" already exists", kubeshellName, desk.Name)
			} else {
				glog.Error(err)
			}
		}
	}
}

func (c *WorkshopController) updateDeskResources(old, new *apiv1.Desk) {
	glog.V(0).Infof("Updating resources for desk \"%s\"", new.Name)
	if old.ResourceVersion == new.ResourceVersion {
		glog.V(0).Infof("No changes changes for desk \"%s\"", new.Name)
		return
	}
	glog.V(0).Infof("Applying changes for desk \"%s\"", new.Name)
}

func (c *WorkshopController) deleteDeskResources(desk *apiv1.Desk) {
	glog.V(0).Infof("Deleting resources for desk \"%s\"", desk.Name)

	trustedNamespaceName := fmt.Sprintf("%s-desk-trusted", desk.Name)
	if err := c.deleteDeskNamespace(desk, trustedNamespaceName); err != nil {
		glog.Errorf("Error deleting namespace \"%s\" for desk \"%s\": %s", trustedNamespaceName, desk.Name, err)
	}

	if err := c.deleteDeskNamespace(desk, fmt.Sprintf("%s-desk-default", desk.Name)); err != nil {
		glog.Errorf("Error deleting namespace \"%s\" for desk \"%s\": %s", trustedNamespaceName, desk.Name, err)
	}
}
