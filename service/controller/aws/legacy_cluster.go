package aws

import (
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
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
// AWSClusterConfig CRD controller implementation.
type LegacyClusterConfig struct {
	ApprClient        *apprclient.Client
	BaseClusterConfig *cluster.Config
	CertSearcher      certs.Interface
	Fs                afero.Fs
	K8sClient         k8sclient.Interface
	Logger            micrologger.Logger
	Tenant            tenantcluster.Interface

	CalicoCIDR        string
	ClusterIPRange    string
	ProjectName       string
	Provider          string
	RegistryDomain    string
	ResourceNamespace string
}

type LegacyCluster struct {
	*controller.Controller
}

// NewLegacyCluster returns a configured AWSClusterConfig controller implementation.
func NewLegacyCluster(config LegacyClusterConfig) (*LegacyCluster, error) {
	var err error

	var resourceSet *controller.ResourceSet
	{
		c := resourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoCIDR:        config.CalicoCIDR,
			ClusterIPRange:    config.ClusterIPRange,
			ProjectName:       config.ProjectName,
			Provider:          config.Provider,
			RegistryDomain:    config.RegistryDomain,
			ResourceNamespace: config.ResourceNamespace,
		}

		resourceSet, err = newResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			CRD:       v1alpha1.NewAWSClusterConfigCRD(),
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSet,
			},
			NewRuntimeObjectFunc: func() pkgruntime.Object {
				return new(v1alpha1.AWSClusterConfig)
			},

			Name: config.ProjectName,
			// ResyncPeriod is 1 minute because some resources access guest
			// clusters. So we need to wait until they become available. When
			// a guest cluster is not available we cancel the reconciliation.
			ResyncPeriod: 1 * time.Minute,
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
