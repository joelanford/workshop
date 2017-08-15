package namespace

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	rbacv1beta1 "k8s.io/client-go/pkg/apis/rbac/v1beta1"
)

type TrustedController struct {
	kubeClient kubernetes.Interface

	serviceAccount v1.ServiceAccount
	namespace      v1.Namespace
	roleBinding    rbacv1beta1.RoleBinding
	resourceQuota  v1.ResourceQuota
}

func NewTrustedController(kubeClient kubernetes.Interface) *TrustedController {
	return &TrustedController{
		kubeClient: kubeClient,
	}
}

func (c *TrustedController) Start() error {
	return nil
}

func (c *TrustedController) Stop() error {
	return nil
}
