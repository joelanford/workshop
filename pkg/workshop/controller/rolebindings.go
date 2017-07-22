package controller

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	rbacv1beta1 "k8s.io/client-go/pkg/apis/rbac/v1beta1"

	"github.com/golang/glog"
	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

func (c *WorkshopController) createDeskRoleBinding(desk *apiv1.Desk, role string, sa *v1.ServiceAccount, namespace *v1.Namespace) (*rbacv1beta1.RoleBinding, error) {
	roleBinding, err := c.kubeClient.RbacV1beta1().RoleBindings(namespace.Name).Create(&rbacv1beta1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", sa.Name, role),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: apiv1.SchemeGroupVersion.String(),
					Kind:       apiv1.DeskKind,
					Name:       desk.Name,
					UID:        desk.UID,
				},
			},
		},
		RoleRef: rbacv1beta1.RoleRef{
			APIGroup: rbacv1beta1.SchemeGroupVersion.Group,
			Kind:     "ClusterRole",
			Name:     role,
		},
		Subjects: []rbacv1beta1.Subject{
			{
				Kind:      rbacv1beta1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: namespace.Name,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("Created rolebinding \"%s\" for serviceaccount \"%s\" in namespace \"%s\" for desk \"%s\"", roleBinding.Name, sa.Name, namespace.Name, desk.Name)

	return roleBinding, nil
}
