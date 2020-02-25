package azure

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/cluster-operator/service/internal/cluster"
)

// LegacyClusterConfig contains necessary dependencies and settings for
// AzureClusterConfig CRD controller implementation.
type LegacyClusterConfig struct {
	ApprClient        *apprclient.Client
	BaseClusterConfig *cluster.Config
	CertSearcher      certs.Interface
	ClusterClient     *clusterclient.Client
	Fs                afero.Fs
	K8sClient         k8sclient.Interface
	Logger            micrologger.Logger
	Tenant            tenantcluster.Interface

	CalicoAddress        string
	CalicoPrefixLength   string
	ClusterIPRange       string
	KubernetesDomain     string
	ProjectName          string
	Provider             string
	RawAppDefaultConfig  string
	RawAppOverrideConfig string
	RegistryDomain       string
	ResourceNamespace    string
}

type LegacyCluster struct {
	*controller.Controller
}

// NewLegacyCluster returns a configured AzureClusterConfig controller implementation.
func NewLegacyCluster(config LegacyClusterConfig) (*LegacyCluster, error) {
	var err error

	var resourceSet *controller.ResourceSet
	{
		c := resourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			ClusterClient:     config.ClusterClient,
			Fs:                config.Fs,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoAddress:        config.CalicoAddress,
			CalicoPrefixLength:   config.CalicoPrefixLength,
			ClusterIPRange:       config.ClusterIPRange,
			KubernetesDomain:     config.KubernetesDomain,
			ProjectName:          config.ProjectName,
			Provider:             config.Provider,
			RawAppDefaultConfig:  config.RawAppDefaultConfig,
			RawAppOverrideConfig: config.RawAppOverrideConfig,
			RegistryDomain:       config.RegistryDomain,
			ResourceNamespace:    config.ResourceNamespace,
		}

		resourceSet, err = newResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			CRD:       v1alpha1.NewAzureClusterConfigCRD(),
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSet,
			},
			NewRuntimeObjectFunc: func() pkgruntime.Object {
				return new(v1alpha1.AzureClusterConfig)
			},

			Name: config.ProjectName,
		}

		clusterController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &LegacyCluster{
		Controller: clusterController,
	}

	return c, nil
}
