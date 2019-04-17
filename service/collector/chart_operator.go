package collector

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"github.com/prometheus/client_golang/prometheus"
)

// ChartOperatorConfig is this collector's configuration struct.
type ChartOperatorConfig struct {
	G8sClient     versioned.Interface
	Helper        *helper
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface
}

// ChartOperator is the main struct for this collector.
type ChartOperator struct {
	g8sClient     versioned.Interface
	helper        *helper
	logger        micrologger.Logger
	tenantCluster tenantcluster.Interface
}

// NewChartOperator creates a new ChartOperator metrics collector
func NewChartOperator(config ChartOperatorConfig) (*ChartOperator, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.TenantCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantCluster must not be empty", config)
	}

	c := &ChartOperator{
		g8sClient:     config.G8sClient,
		logger:        config.Logger,
		tenantCluster: config.TenantCluster,
	}

	return c, nil
}

// Collect is the main metrics collection function.
func (c *ChartOperator) Collect(ch chan<- prometheus.Metric) error {
	// TODO
	return nil
}

// Describe emits the description for the metrics collected here.
func (c *ChartOperator) Describe(ch chan<- *prometheus.Desc) error {
	// TODO
	return nil
}
