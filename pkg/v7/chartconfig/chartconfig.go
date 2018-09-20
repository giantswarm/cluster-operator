package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/kubernetes"
)

const (
	// resourceNamespace is the resource where the chartconfig CRs are created.
	resourceNamespace = "giantswarm"
)

// Config represents the configuration used to create a new chartconfig service.
type Config struct {
	Logger micrologger.Logger
	Tenant tenantcluster.Interface

	ProjectName string
}

// ChartConfig provides shared functionality for managing chartconfigs.
type ChartConfig struct {
	logger micrologger.Logger
	tenant tenantcluster.Interface

	projectName string
}

// New creates a new chartconfig service.
func New(config Config) (*ChartConfig, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Guest must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	s := &ChartConfig{
		logger: config.Logger,
		tenant: config.Tenant,

		projectName: config.ProjectName,
	}

	return s, nil
}

func (c *ChartConfig) newTenantG8sClient(ctx context.Context, clusterConfig ClusterConfig) (versioned.Interface, error) {
	tenantG8sClient, err := c.tenant.NewG8sClient(ctx, clusterConfig.ClusterID, clusterConfig.APIDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tenantG8sClient, nil
}

func (c *ChartConfig) newTenantK8sClient(ctx context.Context, clusterConfig ClusterConfig) (kubernetes.Interface, error) {
	tenantK8sClient, err := c.tenant.NewK8sClient(ctx, clusterConfig.ClusterID, clusterConfig.APIDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tenantK8sClient, nil
}

func getChartConfigByName(list []*v1alpha1.ChartConfig, name string) (*v1alpha1.ChartConfig, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}
