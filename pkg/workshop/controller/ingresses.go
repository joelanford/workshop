package controller

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/golang/glog"
	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

func (c *WorkshopController) createDeskKubeshellIngress(desk *apiv1.Desk, namespace *v1.Namespace, domain string) (*extensionsv1beta1.Ingress, error) {
	deskDomain := fmt.Sprintf("%s.%s", desk.Name, domain)
	ingress, err := c.kubeClient.ExtensionsV1beta1().Ingresses(namespace.Name).Create(&extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeshell",
			Annotations: map[string]string{
				"kubernetes.io/ingress.allow-http":     "false",
				"ingress.kubernetes.io/rewrite-target": "/",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: apiv1.SchemeGroupVersion.String(),
					Kind:       apiv1.DeskKind,
					Name:       desk.Name,
					UID:        desk.UID,
				},
			},
		},
		Spec: extensionsv1beta1.IngressSpec{
			TLS: []extensionsv1beta1.IngressTLS{
				{Hosts: []string{deskDomain}},
			},
			Rules: []extensionsv1beta1.IngressRule{
				{
					Host: deskDomain,
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/kubeshell",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "kubeshell",
										ServicePort: intstr.FromInt(4200),
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("Created ingress \"%s\" in namespace \"%s\" for desk \"%s\"", ingress.Name, namespace.Name, desk.Name)

	return ingress, nil
}
