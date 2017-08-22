package desk

import (
	"k8s.io/client-go/kubernetes"

	"github.com/Sirupsen/logrus"
	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	"github.com/joelanford/workshop/pkg/controller/namespace"
)

type Manager struct {
	kubeClient kubernetes.Interface
	domain     string
	logger     *logrus.Logger

	desk *apiv1.Desk

	trustedNsController *namespace.TrustedController
	defaultNsController *namespace.DefaultController
}

func NewManager(kubeClient kubernetes.Interface, domain string, logger *logrus.Logger, desk *apiv1.Desk) *Manager {
	return &Manager{
		kubeClient:          kubeClient,
		domain:              domain,
		logger:              logger,
		desk:                desk,
		trustedNsController: namespace.NewTrustedController(kubeClient, domain, logger, desk),
		defaultNsController: namespace.NewDefaultController(kubeClient, logger, desk),
	}
}

func (m *Manager) Start() {
	m.logger.Infof("starting desk manager for desk \"%s\"...", m.desk.Name)
	m.trustedNsController.Start()
	m.defaultNsController.Start()
	m.logger.Infof("started desk manager for desk \"%s\"", m.desk.Name)
}

func (m *Manager) Update(d *apiv1.Desk) {
	m.logger.Infof("updating desk \"%s\"...", m.desk.Name)
	m.logger.Infof("updated desk \"%s\"", m.desk.Name)
}

func (m *Manager) Stop() {
	m.logger.Infof("stopping desk manager for desk \"%s\"...", m.desk.Name)
	m.trustedNsController.Stop()
	m.defaultNsController.Stop()
	m.logger.Infof("stopped desk manager for desk \"%s\"", m.desk.Name)
}
