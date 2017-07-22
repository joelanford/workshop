package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/golang/glog"
	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

func (c *WorkshopController) createDeskKubeshellDeployment(desk *apiv1.Desk, inNamespace *v1.Namespace, kubectlNamespace *v1.Namespace) (*extensionsv1beta1.Deployment, error) {
	replicas := int32(1)
	kubeshellLabels := map[string]string{
		"app": "kubeshell",
	}
	deployment, err := c.kubeClient.ExtensionsV1beta1().Deployments(inNamespace.Name).Create(&extensionsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kubeshell",
			Labels: kubeshellLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: apiv1.SchemeGroupVersion.String(),
					Kind:       apiv1.DeskKind,
					Name:       desk.Name,
					UID:        desk.UID,
				},
			},
		},
		Spec: extensionsv1beta1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: kubeshellLabels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: kubeshellLabels,
				},
				Spec: v1.PodSpec{
					ServiceAccountName: desk.Name,
					Containers: []v1.Container{
						{
							Name:  "kubeshell",
							Image: "joelanford/kubeshell:v1.7.0-latest",
							Env: []v1.EnvVar{
								{Name: "KS_USER", Value: desk.Spec.Owner},
								{Name: "KS_IN_CLUSTER", Value: "true"},
								{Name: "KS_NAMESPACE", Value: kubectlNamespace.Name},
								{Name: "KS_ENABLE_SUDO", Value: "false"},
							},
							Ports: []v1.ContainerPort{
								{Protocol: v1.ProtocolTCP, ContainerPort: 4200},
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
	glog.V(1).Infof("Created deployment \"%s\" in namespace \"%s\" for desk \"%s\"", deployment.Name, inNamespace.Name, desk.Name)

	return deployment, nil
}
