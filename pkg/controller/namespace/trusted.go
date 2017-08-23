package namespace

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	kcache "k8s.io/client-go/tools/cache"

	apiv1 "github.com/joelanford/workshop/pkg/apis/workshop/v1"
)

type TrustedController struct {
	kubeClient kubernetes.Interface
	domain     string
	logger     *logrus.Logger

	name string
	desk *apiv1.Desk

	cacheStore      kcache.Store
	cacheController kcache.Controller
	stopCh          chan struct{}

	//	serviceAccount v1.ServiceAccount
	//	roleBinding    rbacv1beta1.RoleBinding
	//	resourceQuota  v1.ResourceQuota

}

func NewTrustedController(kubeClient kubernetes.Interface, domain string, logger *logrus.Logger, desk *apiv1.Desk) *TrustedController {
	c := &TrustedController{
		kubeClient: kubeClient,
		domain:     domain,
		logger:     logger,
		name:       fmt.Sprintf("desk-%s-trusted", desk.Name),
		desk:       desk,
		stopCh:     make(chan struct{}, 0),
	}
	c.cacheStore, c.cacheController = kcache.NewInformer(
		kcache.NewListWatchFromClient(
			kubeClient.CoreV1().RESTClient(),
			"namespaces",
			v1.NamespaceAll,
			fields.OneTermEqualSelector(api.ObjectNameField, c.name),
		),
		&v1.Namespace{},
		5*time.Minute,
		kcache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		},
	)
	return c
}

func (c *TrustedController) Start() error {
	_, err := c.create()
	if err != nil {
		return fmt.Errorf("could not create trusted namespace for desk %s: %s", c.desk.Name, err)
	}

	go c.cacheController.Run(c.stopCh)

	// Wait until the cache controller has synced.
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for !c.cacheController.HasSynced() {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for cache initialization")
		case <-ticker.C:
			c.logger.Info("waiting for trusted namespace cache to be initialized from apiserver...")
		}
	}
	c.logger.Info("successfully initialized namespace cache")
	return nil
}

func (c *TrustedController) Stop() error {
	close(c.stopCh)
	return c.delete()
}

func (c *TrustedController) create() (*v1.Namespace, error) {
	namespace, err := c.kubeClient.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.name,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: apiv1.SchemeGroupVersion.String(),
					Kind:       apiv1.DeskKind,
					Name:       c.desk.Name,
					UID:        c.desk.UID,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	c.logger.Infof("created namespace \"%s\" for desk \"%s\"", namespace.Name, c.desk.Name)
	return namespace, nil
}

func (c *TrustedController) delete() error {
	err := c.kubeClient.CoreV1().Namespaces().Delete(c.name, nil)
	if err != nil {
		return err
	}
	c.logger.Infof("deleted namespace \"%s\" for desk \"%s\"", c.name, c.desk.Name)
	return nil
}

func (c *TrustedController) onAdd(obj interface{}) {

}

func (c *TrustedController) onUpdate(oldObj, newObj interface{}) {

}

func (c *TrustedController) onDelete(obj interface{}) {
	if _, err := c.create(); err != nil {
		c.logger.Errorf("could not re-create trusted namespace for desk %s: %s", c.desk.Name, err)
	}
}
