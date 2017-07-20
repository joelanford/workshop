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

package workshop

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workshopv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	clientv1 "github.com/joelanford/workshop/pkg/client/v1"
	"github.com/joelanford/workshop/pkg/controller/desk"
)

type WorkshopController struct {
	logger    log.Logger
	client    *clientv1.Client
	startTime time.Time

	deskController desk.DeskController
}

func NewController(logger log.Logger, client *clientv1.Client) *WorkshopController {
	return &WorkshopController{
		logger:         logger,
		client:         client,
		deskController: desk.NewController(logger, client),
	}
}

// Run starts a workshop resource controller
func (c *WorkshopController) Run(ctx context.Context, expirationInterval time.Duration) error {
	c.startTime = time.Now()

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
	c.client.WatchDesks(ctx, c.onAdd, c.onUpdate, c.onDelete)
	c.logger.Log("msg", "started watching for desk resource changes")

	// Clean up expired Desk objects
	go c.deskController.WatchExpirations(ctx, expirationInterval)
	c.logger.Log("msg", "started watching for expired desk resources")

	<-ctx.Done()
	c.logger.Log("msg", "stopped watching for desk resource changes")

	return ctx.Err()
}

func (c *WorkshopController) Clean() error {
	return c.client.DeleteDeskCRD()
}

func (c *WorkshopController) onAdd(obj interface{}) {
	switch obj.(type) {
	case *workshopv1.Desk:
		desk := obj.(*workshopv1.Desk).DeepCopyObject().(*workshopv1.Desk)
		if desk.ObjectMeta.CreationTimestamp.Before(metav1.NewTime(c.startTime)) {
			c.logger.Log("msg", "found existing desk", "id", desk.ObjectMeta.UID, "owner", desk.Spec.Owner, "version", desk.Spec.Version, "expiration", desk.Spec.ExpirationTimestamp)
			return
		}
		if err := c.deskController.Prepare(desk); err != nil {
			c.logger.Log("msg", "could not prepare desk for initialization", "msg", err)
		}
		c.logger.Log("msg", "prepared new desk", "id", desk.ObjectMeta.UID, "owner", desk.Spec.Owner, "version", desk.Spec.Version, "expiration", desk.Spec.ExpirationTimestamp)
	default:
		c.logger.Log("msg", "could not handle added object", "err", "unknown type")
	}
}

func (c *WorkshopController) onUpdate(old, new interface{}) {
	switch new.(type) {
	case *workshopv1.Desk:
		oldDesk := old.(*workshopv1.Desk).DeepCopyObject().(*workshopv1.Desk)
		newDesk := new.(*workshopv1.Desk).DeepCopyObject().(*workshopv1.Desk)
		if newDesk.IsStatusUpdate(oldDesk) {
			switch newDesk.Status.State {
			case workshopv1.DeskStatusStateInitializing:
				if err := c.deskController.Initialize(newDesk); err != nil {
					c.logger.Log("msg", "could not initialize desk", "msg", err)
				}
				c.logger.Log("msg", "initialized desk", "id", newDesk.ObjectMeta.UID, "owner", newDesk.Spec.Owner, "version", newDesk.Spec.Version, "expiration", newDesk.Spec.ExpirationTimestamp)
			case workshopv1.DeskStatusStateReady:
				c.logger.Log("msg", "desk is ready", "id", newDesk.ObjectMeta.UID, "owner", newDesk.Spec.Owner, "version", newDesk.Spec.Version, "expiration", newDesk.Spec.ExpirationTimestamp)
			case workshopv1.DeskStatusStateExpired:
				if err := c.deskController.Expire(newDesk); err != nil {
					c.logger.Log("msg", "could not initialize desk", "msg", err)
				}
				c.logger.Log("msg", "expired desk", "id", newDesk.ObjectMeta.UID, "owner", newDesk.Spec.Owner, "version", newDesk.Spec.Version, "expiration", newDesk.Spec.ExpirationTimestamp)
			case workshopv1.DeskStatusStateTerminating:
				c.logger.Log("msg", "desk is terminating", "id", newDesk.ObjectMeta.UID, "owner", newDesk.Spec.Owner, "version", newDesk.Spec.Version, "expiration", newDesk.Spec.ExpirationTimestamp)
			default:
				c.logger.Log("msg", "could not process desk update", "err", "unknown state")
			}
		} else if newDesk.IsSpecUpdate(oldDesk) {
			if err := c.deskController.Update(oldDesk, newDesk); err != nil {
				c.logger.Log("msg", "could not update desk", "msg", err)
			}
			c.logger.Log("msg", "updated desk", "id", newDesk.ObjectMeta.UID, "owner", newDesk.Spec.Owner, "version", newDesk.Spec.Version, "expiration", newDesk.Spec.ExpirationTimestamp)
		}
	default:
		c.logger.Log("msg", "could not handle updated object", "err", "unknown type")
	}
}

func (c *WorkshopController) onDelete(obj interface{}) {
	switch obj.(type) {
	case *workshopv1.Desk:
		desk := obj.(*workshopv1.Desk).DeepCopyObject().(*workshopv1.Desk)
		if err := c.deskController.Terminate(desk); err != nil {
			c.logger.Log("msg", "could not terminate desk", "msg", err)
		}
		c.logger.Log("msg", "terminated desk", "id", desk.ObjectMeta.UID, "owner", desk.Spec.Owner, "version", desk.Spec.Version, "expiration", desk.Spec.ExpirationTimestamp)
	default:
		c.logger.Log("msg", "could not handle deleted object", "err", "unknown type")
	}
}
