package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"

	"github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

type WorkshopV1Interface interface {
	RESTClient() rest.Interface
	DesksGetter
}

type WorkshopV1Client struct {
	restClient rest.Interface
}

func (c *WorkshopV1Client) Desks() DeskInterface {
	return newDesks(c)
}

func NewForConfig(c *rest.Config) (*WorkshopV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &WorkshopV1Client{client}, nil
}

func NewForConfigOrDie(c *rest.Config) *WorkshopV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

func New(c rest.Interface) *WorkshopV1Client {
	return &WorkshopV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		return err
	}

	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

func (c *WorkshopV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
