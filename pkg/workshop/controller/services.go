package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/golang/glog"
	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

func (c *WorkshopController) createDeskKubeshellService(desk *apiv1.Desk, namespace *v1.Namespace) (*v1.Service, error) {
	kubeshellLabels := map[string]string{
		"app": "kubeshell",
	}

	service, err := c.kubeClient.CoreV1().Services(namespace.Name).Create(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeshell",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: apiv1.SchemeGroupVersion.String(),
					Kind:       apiv1.DeskKind,
					Name:       desk.Name,
					UID:        desk.UID,
				},
			},
		},
		Spec: v1.ServiceSpec{
			Selector: kubeshellLabels,
			Ports: []v1.ServicePort{
				{Protocol: v1.ProtocolTCP, Port: 4200, TargetPort: intstr.FromInt(4200)},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("Created service \"%s\" in namespace \"%s\" for desk \"%s\"", service.Name, namespace.Name, desk.Name)

	return service, nil
}
