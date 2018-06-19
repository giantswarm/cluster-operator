package network

import (
	"net"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/service/controller/network/v1"
)

// ControllerConfig contains necessary dependencies and settings for
// ClusterNetworkConfig CRD controller implementation.
type ControllerConfig struct {
	BaseClusterConfig *cluster.Config
	G8sClient         versioned.Interface
	K8sClient         kubernetes.Interface
	K8sExtClient      apiextensionsclient.Interface
	Logger            micrologger.Logger

	Network         net.IPNet
	ProjectName     string
	ReservedSubnets []net.IPNet
	StorageKind     string
}

type Controller struct {
	*controller.Controller
}

// NewController returns a configured ClusterNetworkConfig controller implementation.
func NewController(config ControllerConfig) (*Controller, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}

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
			Watcher: config.G8sClient.CoreV1alpha1().ClusterNetworkConfigs(""),
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v1ResourceSet *controller.ResourceSet
	{
		c := v1.ResourceSetConfig{
			CRDClient: crdClient,
			G8sClient: config.G8sClient,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			Network:         config.Network,
			ProjectName:     config.ProjectName,
			ReservedSubnets: config.ReservedSubnets,
			StorageKind:     config.StorageKind,
		}

		v1ResourceSet, err = v1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceRouter *controller.ResourceRouter
	{
		c := controller.ResourceRouterConfig{
			Logger: config.Logger,
			ResourceSets: []*controller.ResourceSet{
				v1ResourceSet,
			},
		}

		resourceRouter, err = controller.NewResourceRouter(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var networkController *controller.Controller
	{
		c := controller.Config{
			CRD:            v1alpha1.NewClusterNetworkConfigCRD(),
			CRDClient:      crdClient,
			Informer:       newInformer,
			Logger:         config.Logger,
			ResourceRouter: resourceRouter,
			RESTClient:     config.G8sClient.CoreV1alpha1().RESTClient(),

			Name: config.ProjectName,
		}

		networkController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Controller{
		Controller: networkController,
	}

	return c, nil
}
