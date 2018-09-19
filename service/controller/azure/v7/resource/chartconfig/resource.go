package chartconfig

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/v7/chartconfig"
)

const (
	// Name is the identifier of the resource.
	Name = "chartconfigv7"
)

// Config represents the configuration used to create a new chartconfig resource.
type Config struct {
	ChartConfig chartconfig.Interface
	K8sClient   kubernetes.Interface
	Logger      micrologger.Logger

	ProjectName string
}

// Resource implements the chart config resource.
type Resource struct {
	chartConfig chartconfig.Interface
	k8sClient   kubernetes.Interface
	logger      micrologger.Logger

	projectName string
}

// New creates a new configured chartconfig resource.
func New(config Config) (*Resource, error) {
	if config.ConfigMap == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ChartConfig must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	r := &Resource{
		chartConfig: config.ChartConfig,
		k8sClient:   config.K8sClient,
		logger:      config.Logger,

		projectName: config.ProjectName,
	}

	return r, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
}
