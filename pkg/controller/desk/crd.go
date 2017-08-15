package desk

import (
	"fmt"
	"time"

	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

var crd apiextensionsv1beta1.CustomResourceDefinition = apiextensionsv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: apiv1.DeskCRDName,
	},
	Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
		Group:   apiv1.GroupName,
		Version: apiv1.SchemeGroupVersion.Version,
		Scope:   apiextensionsv1beta1.ClusterScoped,
		Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
			Plural: apiv1.DeskResourcePlural,
			Kind:   apiv1.DeskKind,
		},
	},
}

type CRDManager struct {
	apiExtClient apiextensionsclient.Interface
}

func NewCRDManager(apiExtClient apiextensionsclient.Interface) *CRDManager {
	return &CRDManager{
		apiExtClient: apiExtClient,
	}
}

func (m *CRDManager) Create() error {
	_, err := m.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&crd)
	return err
}

func (m *CRDManager) Wait() error {
	// wait for CRD being established
	err := wait.Poll(500*time.Millisecond, 5*time.Second, func() (bool, error) {
		crd, err := m.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(apiv1.DeskCRDName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, nil
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					return true, fmt.Errorf("desk custom resource definition name conflict: %v\n", cond.Reason)
				}
			}
		}
		return false, nil
	})

	if err != nil {
		deleteErr := m.Delete()
		if deleteErr != nil {
			return errors.NewAggregate([]error{err, deleteErr})
		}
		return err
	}
	return nil
}

func (m *CRDManager) CreateAndWait() error {
	if err := m.Create(); err != nil {
		return err
	}
	return m.Wait()
}

func (m *CRDManager) Delete() error {
	return m.apiExtClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(apiv1.DeskCRDName, nil)
}
