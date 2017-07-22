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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"

	"github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

type DesksGetter interface {
	Desks() DeskInterface
}

type DeskInterface interface {
	Create(*v1.Desk) (*v1.Desk, error)
	Update(*v1.Desk) (*v1.Desk, error)
	UpdateStatus(*v1.Desk) (*v1.Desk, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.Desk, error)
	List(opts metav1.ListOptions) (*v1.DeskList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Desk, err error)
	DeskExpansion
}

// desks implements DeskInterface
type desks struct {
	client rest.Interface
}

// newDesks returns a Desks
func newDesks(c *WorkshopV1Client) *desks {
	return &desks{
		client: c.RESTClient(),
	}
}

// Create takes the representation of a desk and creates it.  Returns the server's representation of the desk, and an error, if there is any.
func (c *desks) Create(desk *v1.Desk) (result *v1.Desk, err error) {
	result = &v1.Desk{}
	err = c.client.Post().
		Resource("desks").
		Body(desk).
		Do().
		Into(result)
	return
}

// Update takes the representation of a desk and updates it. Returns the server's representation of the desk, and an error, if there is any.
func (c *desks) Update(desk *v1.Desk) (result *v1.Desk, err error) {
	result = &v1.Desk{}
	err = c.client.Put().
		Resource("desks").
		Name(desk.Name).
		Body(desk).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclientstatus=false comment above the type to avoid generating UpdateStatus().

func (c *desks) UpdateStatus(desk *v1.Desk) (result *v1.Desk, err error) {
	result = &v1.Desk{}
	err = c.client.Put().
		Resource("desks").
		Name(desk.Name).
		SubResource("status").
		Body(desk).
		Do().
		Into(result)
	return
}

// Delete takes name of the desk and deletes it. Returns an error if one occurs.
func (c *desks) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("desks").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *desks) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return c.client.Delete().
		Resource("desks").
		VersionedParams(&listOptions, metav1.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the desk, and returns the corresponding desk object, and an error if there is any.
func (c *desks) Get(name string, options metav1.GetOptions) (result *v1.Desk, err error) {
	result = &v1.Desk{}
	err = c.client.Get().
		Resource("desks").
		Name(name).
		VersionedParams(&options, metav1.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Desks that match those selectors.
func (c *desks) List(opts metav1.ListOptions) (result *v1.DeskList, err error) {
	result = &v1.DeskList{}
	err = c.client.Get().
		Resource("desks").
		VersionedParams(&opts, metav1.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested desks.
func (c *desks) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("desks").
		VersionedParams(&opts, metav1.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched desk.
func (c *desks) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Desk, err error) {
	result = &v1.Desk{}
	err = c.client.Patch(pt).
		Resource("desks").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
