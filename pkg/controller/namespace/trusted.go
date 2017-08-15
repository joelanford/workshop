package namespace

import (
	"k8s.io/client-go/pkg/api/v1"
	rbacv1beta1 "k8s.io/client-go/pkg/apis/rbac/v1beta1"
)

type TrustedController struct {
	serviceAccount v1.ServiceAccount
	namespace      v1.Namespace
	roleBinding    rbacv1beta1.RoleBinding
	resourceQuota  v1.ResourceQuota
}
