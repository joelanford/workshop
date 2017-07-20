package desk

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workshopv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	clientv1 "github.com/joelanford/workshop/pkg/client/v1"
)

type DeskController interface {
	Prepare(desk *workshopv1.Desk) error
	Initialize(desk *workshopv1.Desk) error
	Update(old *workshopv1.Desk, new *workshopv1.Desk) error
	Expire(desk *workshopv1.Desk) error
	WatchExpirations(ctx context.Context, expirationInterval time.Duration) error
	Terminate(desk *workshopv1.Desk) error
}

type deskController struct {
	logger log.Logger
	client *clientv1.Client
}

func NewController(logger log.Logger, client *clientv1.Client) *deskController {
	return &deskController{
		logger: logger,
		client: client,
	}
}

func statusUpdateError(err error, state workshopv1.DeskStatusState) error {
	return errors.Wrapf(err, "could not set desk status to %s", state)
}

func (c *deskController) Prepare(desk *workshopv1.Desk) error {
	maxTimestamp := time.Now().UTC().Add(workshopv1.DeskMaxLifespan)

	//
	// TODO: How to return this error condition to the client?
	//       For now, just silently change it to the max.
	//
	if desk.Spec.ExpirationTimestamp.After(maxTimestamp) {
		desk.Spec.ExpirationTimestamp = metav1.NewTime(maxTimestamp)
	}
	if desk.Spec.ExpirationTimestamp.IsZero() {
		desk.Spec.ExpirationTimestamp = metav1.NewTime(maxTimestamp)
	}

	if desk.Spec.Version == "" {
		desk.Spec.Version = "latest"
	}

	desk.Status = workshopv1.DeskStatus{
		State:   workshopv1.DeskStatusStateInitializing,
		Message: workshopv1.DeskStatusMsgInitializing,
	}

	if err := c.client.PutDesk(desk); err != nil {
		return statusUpdateError(err, workshopv1.DeskStatusStateInitializing)
	}
	return nil
}

func (c *deskController) Initialize(desk *workshopv1.Desk) error {
	return nil
}

func (c *deskController) Update(old *workshopv1.Desk, new *workshopv1.Desk) error {
	c.logger.Log("msg", "logging updated spec", "old", fmt.Sprintf("%+v", old.Spec), "new", fmt.Sprintf("%+v", new.Spec))
	return nil
}

func (c *deskController) Expire(desk *workshopv1.Desk) error {
	return nil
}

func (c *deskController) WatchExpirations(ctx context.Context, expirationInterval time.Duration) error {
	ticker := time.NewTicker(expirationInterval)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			c.logger.Log("msg", "checking for expired desks")
			deskList, err := c.client.ListDesks()
			if err != nil {
				return err
			}
			for _, desk := range deskList.Items {
				if desk.Spec.ExpirationTimestamp.Before(metav1.NewTime(now)) &&
					desk.Status.State != workshopv1.DeskStatusStateExpired &&
					desk.Status.State != workshopv1.DeskStatusStateTerminating {
					desk.Status = workshopv1.DeskStatus{
						State:   workshopv1.DeskStatusStateExpired,
						Message: workshopv1.DeskStatusMsgExpired,
					}
					if err := c.client.PutDesk(&desk); err != nil {
						return statusUpdateError(err, workshopv1.DeskStatusStateExpired)
					}
					c.logger.Log("msg", "updated desk for expiration", "id", desk.ObjectMeta.UID, "owner", desk.Spec.Owner, "version", desk.Spec.Version, "expiration", desk.Spec.ExpirationTimestamp)
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *deskController) Terminate(desk *workshopv1.Desk) error {
	return nil
}
