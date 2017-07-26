package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/golang/glog"
	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

func (c *WorkshopController) createDeskNamespace(desk *apiv1.Desk, name string) (*v1.Namespace, error) {
	namespace, err := c.kubeClient.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
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
	glog.V(1).Infof("Created namespace \"%s\" for desk \"%s\"", namespace.Name, desk.Name)
	return namespace, nil
}

func (c *WorkshopController) deleteDeskNamespace(desk *apiv1.Desk, name string) error {
	err := c.kubeClient.CoreV1().Namespaces().Delete(name, nil)
	if err != nil {
		return err
	}
	glog.V(1).Infof("Deleted namespace \"%s\" for desk \"%s\"", name, desk.Name)

	return nil
}
