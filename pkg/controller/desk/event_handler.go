package desk

import (
	"github.com/Sirupsen/logrus"

	"k8s.io/client-go/kubernetes"

	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	"github.com/joelanford/workshop/pkg/client/workshop"
)

type EventHandler struct {
	kubeClient     kubernetes.Interface
	workshopClient workshop.Interface
	logger         *logrus.Logger
	desks          map[string]Controller
}

func NewEventHandler(kubeClient kubernetes.Interface, workshopClient workshop.Interface, logger *logrus.Logger) *EventHandler {
	return &EventHandler{
		kubeClient:     kubeClient,
		workshopClient: workshopClient,
		logger:         logger,
		desks:          make(map[string]Controller),
	}
}

func (eh *EventHandler) OnAdd(obj interface{}) {
	if d, ok := obj.(*apiv1.Desk); ok {
		_ = d
		deskController := NewController(eh.kubeClient, *d)
		deskController.Start()
		eh.desks[d.Name] = *deskController
	}
}

func (eh *EventHandler) OnUpdate(oldObj, newObj interface{}) {
	if d, ok := newObj.(*apiv1.Desk); ok {
		deskController, ok := eh.desks[d.Name]
		if !ok {
			eh.logger.Errorf("could not delete desk resources for deleted desk \"%s\": no controller found", d.Name)
			return
		}
		deskController.Update(*d)
	}
}

func (eh *EventHandler) OnDelete(obj interface{}) {
	if d, ok := obj.(*apiv1.Desk); ok {
		deskController, ok := eh.desks[d.Name]
		if !ok {
			eh.logger.Errorf("could not delete desk resources for deleted desk \"%s\": no controller found", d.Name)
			return
		}
		deskController.Stop()
	}
}
