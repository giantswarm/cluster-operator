package chartconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-operator/pkg/v14/chartconfig"
)

const (
	// Name is the identifier of the resource.
	Name = "chartconfigv14"
)

// Config represents the configuration used to create a new chartconfig resource.
type Config struct {
	ChartConfig chartconfig.Interface
	Logger      micrologger.Logger

	ProjectName string
}

// Resource implements the chart config resource.
type Resource struct {
	chartConfig chartconfig.Interface
	logger      micrologger.Logger

	projectName string
}

// New creates a new configured chartconfig resource.
func New(config Config) (*Resource, error) {
	if config.ChartConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ChartConfig must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	r := &Resource{
		chartConfig: config.ChartConfig,
		logger:      config.Logger,

		projectName: config.ProjectName,
	}

	return r, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
}

func toChartConfigs(v interface{}) ([]*v1alpha1.ChartConfig, error) {
	if v == nil {
		return nil, nil
	}

	chartConfigs, ok := v.([]*v1alpha1.ChartConfig)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*v1alpha1.ChartConfig{}, v)
	}

	return chartConfigs, nil
}
