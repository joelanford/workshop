package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/golang/glog"
	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

func (c *WorkshopController) createDeskServiceAccount(desk *apiv1.Desk, namespace *v1.Namespace) (*v1.ServiceAccount, error) {
	sa, err := c.kubeClient.CoreV1().ServiceAccounts(namespace.Name).Create(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: desk.Spec.Owner,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: apiv1.SchemeGroupVersion.String(),
					Kind:       apiv1.DeskKind,
					Name:       desk.Name,
					UID:        desk.UID,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("Created serviceaccount \"%s\" in namespace \"%s\" for desk \"%s\"", sa.Name, namespace.Name, desk.Name)

	return sa, nil
}
