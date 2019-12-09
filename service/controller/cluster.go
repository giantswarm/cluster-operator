package controller

import (
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/runtime"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/project"
)

// ClusterConfig contains necessary dependencies and settings for
// Cluster API's Cluster CRD controller implementation.
type ClusterConfig struct {
	ApprClient    *apprclient.Client
	CertsSearcher certs.Interface
	ClusterClient *clusterclient.Client
	FileSystem    afero.Fs
	K8sClient     k8sclient.Interface
	Logger        micrologger.Logger
	Tenant        tenantcluster.Interface

	APIIP              string
	CalicoAddress      string
	CalicoPrefixLength string
	CertTTL            string
	ClusterIPRange     string
	DNSIP              string
	Provider           string
	RegistryDomain     string
}

type Cluster struct {
	*controller.Controller
}

// NewCluster returns a configured AWSClusterConfig controller implementation.
func NewCluster(config ClusterConfig) (*Cluster, error) {
	var err error

	var resourceSet *controller.ResourceSet
	{
		c := clusterResourceSetConfig{
			ApprClient:    config.ApprClient,
			CertsSearcher: config.CertsSearcher,
			ClusterClient: config.ClusterClient,
			FileSystem:    config.FileSystem,
			G8sClient:     config.K8sClient.G8sClient(),
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			Tenant:        config.Tenant,

			APIIP:              config.APIIP,
			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			CertTTL:            config.CertTTL,
			ClusterIPRange:     config.ClusterIPRange,
			DNSIP:              config.DNSIP,
			Provider:           config.Provider,
			RegistryDomain:     config.RegistryDomain,
		}

		resourceSet, err = newClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			CRD:       infrastructurev1alpha2.NewClusterCRD(),
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSet,
			},
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(apiv1alpha2.Cluster)
			},

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/cluster-operator-cluster-controller.
			Name: project.Name() + "-cluster-controller",
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
