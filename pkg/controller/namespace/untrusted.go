package namespace

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	rbacv1beta1 "k8s.io/client-go/pkg/apis/rbac/v1beta1"
)

type UntrustedController struct {
	kubeClient kubernetes.Interface

	serviceAccount v1.ServiceAccount
	namespace      v1.Namespace
	roleBinding    rbacv1beta1.RoleBinding
	resourceQuota  v1.ResourceQuota
}

func NewUntrustedController(kubeClient kubernetes.Interface) *UntrustedController {
	return &UntrustedController{
		kubeClient: kubeClient,
	}
}

func (c *UntrustedController) Start() error {
	return nil
}

func (c *UntrustedController) Stop() error {
	return nil
}
