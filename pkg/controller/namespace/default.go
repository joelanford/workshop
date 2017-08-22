package namespace

import (
	"fmt"

	"github.com/Sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	rbacv1beta1 "k8s.io/client-go/pkg/apis/rbac/v1beta1"

	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

type DefaultController struct {
	kubeClient kubernetes.Interface
	logger     *logrus.Logger

	desk *apiv1.Desk

	serviceAccount v1.ServiceAccount
	namespace      v1.Namespace
	roleBinding    rbacv1beta1.RoleBinding
	resourceQuota  v1.ResourceQuota

	namespaceName string
}

func NewDefaultController(kubeClient kubernetes.Interface, logger *logrus.Logger, desk *apiv1.Desk) *DefaultController {
	return &DefaultController{
		kubeClient:    kubeClient,
		logger:        logger,
		desk:          desk,
		namespaceName: fmt.Sprintf("desk-%s-default", desk.Name),
	}
}

func (c *DefaultController) Start() error {
	_, err := c.create()
	return err
}

func (c *DefaultController) Stop() error {
	return c.delete()
}

func (c *DefaultController) create() (*v1.Namespace, error) {
	namespace, err := c.kubeClient.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.namespaceName,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: apiv1.SchemeGroupVersion.String(),
					Kind:       apiv1.DeskKind,
					Name:       c.desk.Name,
					UID:        c.desk.UID,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	c.logger.Infof("created namespace \"%s\" for desk \"%s\"", namespace.Name, c.desk.Name)
	return namespace, nil
}

func (c *DefaultController) delete() error {
	err := c.kubeClient.CoreV1().Namespaces().Delete(c.namespaceName, nil)
	if err != nil {
		return err
	}
	c.logger.Infof("deleted namespace \"%s\" for desk \"%s\"", c.namespaceName, c.desk.Name)
	return nil
}
