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

	trustedNsController   namespace.TrustedController
	untrustedNsController namespace.UntrustedController
}

func NewManager(kubeClient kubernetes.Interface, domain string, logger *logrus.Logger, desk *apiv1.Desk) *Manager {
	return &Manager{
		kubeClient: kubeClient,
		domain:     domain,
		logger:     logger,
		desk:       desk,
	}
}

func (m *Manager) Start() {
	m.logger.Debugf("starting desk manager for desk \"%s\"...", m.desk.Name)
	m.trustedNsController.Start()
	m.untrustedNsController.Start()
	m.logger.Debugf("started desk manager for desk \"%s\"", m.desk.Name)
}

func (m *Manager) Update(d *apiv1.Desk) {
	m.logger.Debugf("updating desk \"%s\"...", m.desk.Name)
	m.logger.Debugf("updated desk \"%s\"", m.desk.Name)
}

func (m *Manager) Stop() {
	m.logger.Debugf("stopping desk manager for desk \"%s\"...", m.desk.Name)
	m.trustedNsController.Stop()
	m.untrustedNsController.Stop()
	m.logger.Debugf("stopped desk manager for desk \"%s\"", m.desk.Name)
}
