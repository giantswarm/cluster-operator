package clusterapi

import (
	clusterv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	v22 "github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22"
)

type MachineDeploymentConfig struct {
	CMAClient    clientset.Interface
	G8sClient    versioned.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger
	Tenant       tenantcluster.Interface

	ProjectName string
	Provider    string
}

type MachineDeployment struct {
	*controller.Controller
}

func NewMachineDeployment(config MachineDeploymentConfig) (*MachineDeployment, error) {
	var err error

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.Config{
			K8sExtClient: config.K8sExtClient,
			Logger:       config.Logger,
		}

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Logger:  config.Logger,
			Watcher: config.CMAClient.ClusterV1alpha1().MachineDeployments(corev1.NamespaceAll),

			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV22 *controller.ResourceSet
	{
		c := v22.MachineDeploymentResourceSetConfig{
			CMAClient: config.CMAClient,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
			Tenant:    config.Tenant,

			Provider: config.Provider,
		}

		resourceSetV22, err = v22.NewMachineDeploymentResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			CRD:       clusterv1alpha1.NewMachineDeploymentCRD(),
			CRDClient: crdClient,
			Informer:  newInformer,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSetV22,
			},
			RESTClient: config.CMAClient.ClusterV1alpha1().RESTClient(),

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/cluster-operator-machine-deployment-controller.
			Name: config.ProjectName + "-machine-deployment-controller",
		}

		clusterController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &MachineDeployment{
		Controller: clusterController,
	}

	return c, nil
}
