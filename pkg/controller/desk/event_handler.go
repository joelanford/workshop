package desk

import (
	"github.com/Sirupsen/logrus"

	"k8s.io/client-go/kubernetes"

	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

type EventHandler struct {
	kubeClient kubernetes.Interface
	domain     string
	logger     *logrus.Logger
	desks      map[string]Manager
}

func NewEventHandler(kubeClient kubernetes.Interface, domain string, logger *logrus.Logger) *EventHandler {
	return &EventHandler{
		kubeClient: kubeClient,
		domain:     domain,
		logger:     logger,
		desks:      make(map[string]Manager),
	}
}

func (eh *EventHandler) OnAdd(obj interface{}) {
	if d, ok := obj.(*apiv1.Desk); ok {
		desk := d.DeepCopyObject().(*apiv1.Desk)
		deskManager := NewManager(eh.kubeClient, eh.domain, eh.logger, desk)
		deskManager.Start()
		eh.desks[desk.Name] = *deskManager
	}
}

func (eh *EventHandler) OnUpdate(oldObj, newObj interface{}) {
	if d, ok := newObj.(*apiv1.Desk); ok {
		desk := d.DeepCopyObject().(*apiv1.Desk)
		deskManager, ok := eh.desks[desk.Name]
		if !ok {
			eh.logger.Errorf("could not update desk resources for desk \"%s\": no controller found", d.Name)
			return
		}
		deskManager.Update(desk)
	}
}

func (eh *EventHandler) OnDelete(obj interface{}) {
	if d, ok := obj.(*apiv1.Desk); ok {
		deskManager, ok := eh.desks[d.Name]
		if !ok {
			eh.logger.Errorf("could not delete desk resources for desk \"%s\": no controller found", d.Name)
			return
		}
		deskManager.Stop()
		delete(eh.desks, d.Name)
	}
}
