package desk

import (
	"k8s.io/client-go/kubernetes"

	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	"github.com/joelanford/workshop/pkg/controller/namespace"
)

type Controller struct {
	kubeClient kubernetes.Interface

	desk apiv1.Desk

	namespaceCache namespaceCache

	trustedNsController   namespace.TrustedController
	untrustedNsController namespace.UntrustedController
}

type namespaceCache struct {
}

func NewController(kubeClient kubernetes.Interface, desk apiv1.Desk) *Controller {
	return &Controller{
		kubeClient: kubeClient,
		desk:       desk,
	}
}

func (c *Controller) Start() {

}

func (c *Controller) Update(d apiv1.Desk) {

}

func (c *Controller) Stop() {

}
