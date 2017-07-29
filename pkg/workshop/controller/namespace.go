package controller

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/pkg/api/v1"
	kcache "k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

func (c *WorkshopController) setNamespacesStore() {
	// Returns a cache.ListWatch that gets all changes to namespaces.
	c.namespacesStore, c.namespacesController = kcache.NewInformer(
		kcache.NewListWatchFromClient(
			c.kubeClient.CoreV1().RESTClient(),
			"namespaces",
			v1.NamespaceAll,
			fields.Everything()),
		&v1.Namespace{},
		resyncPeriod,
		kcache.ResourceEventHandlerFuncs{
			DeleteFunc: c.handleNamespaceDelete,
		},
	)
}

func (c *WorkshopController) waitForNamespacesSynced() error {
	// Wait for controllers to have completed an initial resource listing
	timeout := time.After(c.initialSyncTimeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timeout waiting for initialization")
		case <-ticker.C:
			if c.namespacesController.HasSynced() {
				glog.V(0).Infof("Initialized namespaces from apiserver")
				return nil
			}
			glog.V(0).Infof("Waiting for namespaces to be initialized from apiserver...")
		}
	}
}

func (c *WorkshopController) handleNamespaceDelete(obj interface{}) {
	if namespace, ok := obj.(*v1.Namespace); ok {
		for _, ownerRef := range namespace.OwnerReferences {
			if ownerRef.Kind == apiv1.DeskKind {
				if deskObj, exists, _ := c.desksStore.GetByKey(ownerRef.Name); exists {
					if desk, ok := deskObj.(*apiv1.Desk); ok {
						// The desk still exists, so recreate the namespace
						c.createDeskNamespace(desk, namespace.Name)
					}
				}
			}
		}
	}
}

func (c *WorkshopController) createDeskNamespace(desk *apiv1.Desk, name string) (*v1.Namespace, error) {
	namespace, err := c.kubeClient.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: apiv1.SchemeGroupVersion.String(),
					Kind:       apiv1.DeskKind,
					Name:       desk.Name,
					UID:        desk.UID,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("Created namespace \"%s\" for desk \"%s\"", namespace.Name, desk.Name)
	return namespace, nil
}

func (c *WorkshopController) deleteDeskNamespace(desk *apiv1.Desk, name string) error {
	err := c.kubeClient.CoreV1().Namespaces().Delete(name, nil)
	if err != nil {
		return err
	}
	glog.V(1).Infof("Deleted namespace \"%s\" for desk \"%s\"", name, desk.Name)

	return nil
}
