/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	workshopv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

type Client struct {
	apiextensionsclientset *apiextensionsclient.Clientset
	restClient             *rest.RESTClient
	scheme                 *runtime.Scheme
}

func New(conf *rest.Config) (*Client, error) {
	scheme := runtime.NewScheme()
	if err := workshopv1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	config := *conf
	config.GroupVersion = &workshopv1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	restClient, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	apiextensionsclientset, err := apiextensionsclient.NewForConfig(&config)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create API extensions client from config")
	}

	return &Client{
		apiextensionsclientset: apiextensionsclientset,
		restClient:             restClient,
		scheme:                 scheme,
	}, nil
}

func NewFromFile(kubeconfig string) (*Client, error) {
	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	conf, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not load config from file")
	}

	return New(conf)
}

func NewFromCluster() (*Client, error) {
	conf, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "could not load config from cluster environment")
	}
	return New(conf)
}

func (c *Client) CopyObject(in runtime.Object) (runtime.Object, error) {
	return c.scheme.Copy(in)
}
