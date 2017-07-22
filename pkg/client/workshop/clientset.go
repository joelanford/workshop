package workshop

import (
	workshopv1 "github.com/joelanford/workshop/pkg/client/workshop/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	WorkshopV1() workshopv1.WorkshopV1Interface
}

type Clientset struct {
	*workshopv1.WorkshopV1Client
}

func (c *Clientset) WorkshopV1() workshopv1.WorkshopV1Interface {
	if c == nil {
		return nil
	}
	return c.WorkshopV1Client
}

func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}

	var cs Clientset
	var err error
	cs.WorkshopV1Client, err = workshopv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.WorkshopV1Client = workshopv1.NewForConfigOrDie(c)
	return &cs
}

func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.WorkshopV1Client = workshopv1.New(c)
	return &cs
}
