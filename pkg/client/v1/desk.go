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
	"context"
	"time"

	workshopv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

func (c *Client) CreateDesk(name, owner, version string, expirationTimestamp time.Time) (*workshopv1.Desk, error) {
	desk := &workshopv1.Desk{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: workshopv1.DeskSpec{
			Version:             version,
			Owner:               owner,
			ExpirationTimestamp: metav1.NewTime(expirationTimestamp),
			// AutoUpdate: true|false,
			// Environment: dev|integration|staging|production,
		},
		Status: workshopv1.DeskStatus{
			State:   workshopv1.DeskStateInitializing,
			Message: "Initializing, not assigned yet",
		},
	}
	var result workshopv1.Desk
	err := c.restClient.Post().
		Resource(workshopv1.DeskResourcePlural).
		Body(desk).
		Do().Into(&result)
	if err != nil {
		return nil, err
	}

	if err := c.WaitDesk(name, workshopv1.DeskStateAssigned); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetDesk(name string) (*workshopv1.Desk, error) {
	var result workshopv1.Desk
	err := c.restClient.Get().
		Resource(workshopv1.DeskResourcePlural).
		Name(name).
		Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListDesks() (*workshopv1.DeskList, error) {
	var result workshopv1.DeskList
	err := c.restClient.Get().
		Resource(workshopv1.DeskResourcePlural).
		Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteDesk(name string) error {
	return c.restClient.Delete().
		Resource(workshopv1.DeskResourcePlural).
		Name(name).
		Do().Error()

	// TODO: wait for deletion to occur before returning
}

func (c *Client) DeleteAllDesks() error {
	return c.restClient.Delete().
		Resource(workshopv1.DeskResourcePlural).
		Do().Error()

	// TODO: how do I know what to log?
	// TODO: wait for deletion to occur before returning
}

func (c *Client) PutDesk(desk *workshopv1.Desk) error {
	return c.restClient.Put().
		Name(desk.ObjectMeta.Name).
		Resource(workshopv1.DeskResourcePlural).
		Body(desk).
		Do().
		Error()
}

func (c *Client) WaitDesk(name string, state workshopv1.DeskState) error {
	return wait.PollImmediate(100*time.Millisecond, 10*time.Second, func() (bool, error) {
		var desk workshopv1.Desk
		err := c.restClient.Get().
			Resource(workshopv1.DeskResourcePlural).
			Name(name).
			Do().Into(&desk)

		if err == nil && desk.Status.State == state {
			return true, nil
		}

		return false, err
	})
}

func (c *Client) WatchDesks(ctx context.Context, onAdd func(interface{}), onUpdate func(interface{}, interface{}), onDelete func(interface{}),
) (cache.Controller, error) {
	source := cache.NewListWatchFromClient(
		c.restClient,
		workshopv1.DeskResourcePlural,
		apiv1.NamespaceAll,
		fields.Everything())

	_, controller := cache.NewInformer(
		source,

		// The object type.
		&workshopv1.Desk{},

		// resyncPeriod
		// Every resyncPeriod, all resources in the cache will retrigger events.
		// Set to 0 to disable the resync.
		0,

		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			AddFunc:    onAdd,
			UpdateFunc: onUpdate,
			DeleteFunc: onDelete,
		})

	go controller.Run(ctx.Done())
	return controller, nil
}
