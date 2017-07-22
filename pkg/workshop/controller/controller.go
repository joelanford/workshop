package controller

import (
	"context"
	"time"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	kcache "k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
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
	domain string

	kubeClient     kubernetes.Interface
	apiExtClient   apiextensionsclient.Interface
	workshopClient workshop.Interface

	desksStore      kcache.Store
	desksController kcache.Controller
	desksCache      map[string]*apiv1.Desk

	initialSyncTimeout time.Duration
}

func NewWorkshopController(domain string, kubeClient kubernetes.Interface, apiExtClient apiextensionsclient.Interface, workshopClient workshop.Interface, timeout time.Duration) *WorkshopController {
	c := &WorkshopController{
		domain: domain,

		kubeClient:     kubeClient,
		apiExtClient:   apiExtClient,
		workshopClient: workshopClient,

		desksCache: make(map[string]*apiv1.Desk),

		initialSyncTimeout: timeout,
	}
	c.setDesksStore()
	return c
}

func (c *WorkshopController) Start(ctx context.Context) error {
	glog.V(2).Infof("Creating desk custom resource definitions")
	if err := c.createDeskCRD(); err != nil {
		if apierrors.IsAlreadyExists(err) {
			glog.V(1).Infoln("Desk custom resource definition already exists, continuing execution")
		} else {
			glog.Fatalf("Could not create Desk custom resource definition: %v", err)
		}
	}
	glog.V(2).Infof("Starting desksController")
	go c.desksController.Run(ctx.Done())

	// Wait synchronously for the initial list operations to be
	// complete of desks from APIServer.
	return c.waitForDesksSynced()
}

func (c *WorkshopController) Clean() error {
	glog.Infof("Cleaning deskController resources")
	return c.deleteDeskCRD()
}
