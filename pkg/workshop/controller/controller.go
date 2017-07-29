package controller

import (
	"context"
	"os"
	"path/filepath"
	"time"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	workshopv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	"github.com/joelanford/workshop/pkg/client/workshop"
)

//
// TODO:
//   - Create a new struct type to store shared desk information.
//   - Automatically re-create deleted resources owned by a desk
//     (e.g. namespace, role binding, serviceaccount, deployment, etc.)
//

const (
	// Resync period for the kube controller loop.
	resyncPeriod = 5 * time.Minute
)

type WorkshopController struct {
	domain             string
	initialSyncTimeout time.Duration

	kubeClient     kubernetes.Interface
	apiExtClient   apiextensionsclient.Interface
	workshopClient workshop.Interface

	desksStore      kcache.Store
	desksController kcache.Controller

	namespacesStore      kcache.Store
	namespacesController kcache.Controller
}

func NewWorkshopController(kubeconfig string, domain string, timeout time.Duration) (*WorkshopController, error) {
	c := &WorkshopController{
		domain:             domain,
		initialSyncTimeout: timeout,
	}
	if err := c.setClients(kubeconfig); err != nil {
		return nil, err
	}
	c.setDesksStore()
	c.setNamespacesStore()
	return c, nil
}

func (c *WorkshopController) Start(ctx context.Context) error {
	glog.V(1).Infof("Creating desk custom resource definitions")
	if err := c.createDeskCRD(); err != nil {
		if apierrors.IsAlreadyExists(err) {
			glog.V(1).Infoln("Desk custom resource definition already exists, continuing execution")
		} else {
			glog.Fatalf("Could not create Desk custom resource definition: %v", err)
		}
	}

	c.cleanStaleResources()

	glog.V(2).Infof("Starting desksController")
	go c.desksController.Run(ctx.Done())
	go c.namespacesController.Run(ctx.Done())

	// TODO: Refactor and improve this to wait concurrently for all resources
	// to sync.
	//
	// Wait synchronously for the initial list operations to be
	// complete of desks from APIServer.
	c.waitForDesksSynced()
	return c.waitForNamespacesSynced()
}

func (c *WorkshopController) Clean() error {
	glog.V(1).Infof("Deleting desk custom resource definition")
	return c.deleteDeskCRD()
}

func (c *WorkshopController) setClients(kubeconfig string) error {
	var (
		config *rest.Config
		err    error
	)

	defaultKubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else if _, err := os.Stat(defaultKubeconfig); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
	} else {
		if config, err = rest.InClusterConfig(); err != nil {
			config, _ = clientcmd.BuildConfigFromFlags("http://localhost:8080", "")
		}
	}

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

func (c *WorkshopController) cleanStaleResources() {
	glog.V(1).Infof("Cleaning up desk resources whose desk is no longer present...")
	nsList, err := c.kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("Error listing namespaces for staleness check and removal: %s", err)
	}

	deskList, err := c.workshopClient.WorkshopV1().Desks().List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("Error listing desks for staleness check and removal: %s", err)
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
						glog.Errorf("Error removing stale namespace \"%s\": %s", namespace.Name, err)
						break ownerRefsLoop
					}
					glog.V(0).Infof("Removed stale namespace \"%s\", which was owned by non-existent desk \"%s\"", namespace.Name, ownerRef.Name)
				}
			}
		}
	}
}
