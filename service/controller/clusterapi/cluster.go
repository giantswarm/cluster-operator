package clusterapi

import (
	clusterv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	v17 "github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17"
)

// ClusterConfig contains necessary dependencies and settings for
// Cluster API's Cluster CRD controller implementation.
type ClusterConfig struct {
	BaseClusterConfig *cluster.Config
	ClusterClient     *clusterclient.Client
	CMAClient         clientset.Interface
	G8sClient         versioned.Interface
	K8sExtClient      apiextensionsclient.Interface
	Logger            micrologger.Logger
	Tenant            tenantcluster.Interface

	ProjectName string
}

type Cluster struct {
	*controller.Controller
}

// NewCluster returns a configured AWSClusterConfig controller implementation.
func NewCluster(config ClusterConfig) (*Cluster, error) {
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
			Watcher: config.CMAClient.ClusterV1alpha1().Clusters(corev1.NamespaceAll),

			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV17 *controller.ResourceSet
	{
		c := v17.ClusterResourceSetConfig{
			BaseClusterConfig: config.BaseClusterConfig,
			ClusterClient:     config.ClusterClient,
			CMAClient:         config.CMAClient,
			G8sClient:         config.G8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,
		}

		resourceSetV17, err = v17.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			CRD:       clusterv1alpha1.NewClusterCRD(),
			CRDClient: crdClient,
			Informer:  newInformer,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSetV17,
			},
			RESTClient: config.CMAClient.ClusterV1alpha1().RESTClient(),

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/cluster-operator-cluster-controller.
			Name: config.ProjectName + "-cluster-controller",
		}

		clusterController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Cluster{
		Controller: clusterController,
	}

	return c, nil
}
