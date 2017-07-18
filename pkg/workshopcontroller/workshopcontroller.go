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

package workshopcontroller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workshopv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	clientv1 "github.com/joelanford/workshop/pkg/client/v1"
)

type WorkshopController struct {
	logger log.Logger
	client *clientv1.Client
}

func New(logger log.Logger, client *clientv1.Client) *WorkshopController {
	return &WorkshopController{
		logger: logger,
		client: client,
	}
}

// Run starts an Desk resource controller
func (c *WorkshopController) Run(ctx context.Context) error {
	if _, err := c.client.CreateDeskCRD(); err != nil {
		if apierrors.IsAlreadyExists(err) {
			c.logger.Log("msg", "custom resource definition already exists, continuing execution")
		} else {
			return errors.Wrapf(err, "could create custom resource definition")
		}
	} else {
		c.logger.Log("msg", "successfully created custom resource definition")
	}

	// Watch Desk objects
	_, err := c.client.WatchDesks(ctx, c.onAdd, c.onUpdate, c.onDelete)
	if err != nil {
		return errors.Wrapf(err, "failed to register watch for desk resources")
	}
	c.logger.Log("msg", "started watching for desk resource changes")

	<-ctx.Done()
	c.logger.Log("msg", "stopped watching for desk resource changes")

	return ctx.Err()
}

func (c *WorkshopController) Clean() error {
	return c.client.DeleteDeskCRD()
}

func (c *WorkshopController) onAdd(obj interface{}) {
	desk := obj.(*workshopv1.Desk)
	c.logger.Log("msg", "desk added", "id", desk.ObjectMeta.UID, "owner", desk.Spec.Owner)

	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use c.client.CopyDesk() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	copyObj, err := c.client.CopyObject(desk)
	if err != nil {
		c.logger.Log("msg", "could not create copy of desk object - desk resources not provisioned", "err", err)
		return
	}

	deskCopy := copyObj.(*workshopv1.Desk)

	if deskCopy.Spec.Version == "" {
		deskCopy.Spec.Version = "latest"
	}
	if deskCopy.Spec.ExpirationTimestamp.IsZero() {
		deskCopy.Spec.ExpirationTimestamp = metav1.NewTime(time.Now().Add(time.Hour * 24 * 14).UTC())
	}

	deskCopy.Status = workshopv1.DeskStatus{
		State:   workshopv1.DeskStateAssigned,
		Message: "Successfully assigned by workshop-controller",
	}

	if err := c.client.PutDesk(deskCopy); err != nil {
		c.logger.Log("msg", "could not update desk status", "err", err)
	}
}

func (c *WorkshopController) onUpdate(oldObj, newObj interface{}) {
	oldDesk := oldObj.(*workshopv1.Desk)
	newDesk := newObj.(*workshopv1.Desk)
	c.logger.Log("msg", "desk updated", "old", fmt.Sprintf("%+v", oldDesk), "new", fmt.Sprintf("%+v", newDesk))
}

func (c *WorkshopController) onDelete(obj interface{}) {
	desk := obj.(*workshopv1.Desk)
	c.logger.Log("msg", "desk deleted", "id", desk.ObjectMeta.UID, "owner", desk.Spec.Owner)
}
