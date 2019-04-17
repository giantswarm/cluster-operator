package collector

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/cluster-operator/service/collector/key"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

// ChartOperatorConfig is this collector's configuration struct.
type ChartOperatorConfig struct {
	G8sClient versioned.Interface
	Helper    *helper
	Logger    micrologger.Logger
	Tenant    tenantcluster.Interface
}

// ChartOperator is the main struct for this collector.
type ChartOperator struct {
	g8sClient versioned.Interface
	helper    *helper
	logger    micrologger.Logger
	tenant    tenantcluster.Interface
}

// NewChartOperator creates a new ChartOperator metrics collector
func NewChartOperator(config ChartOperatorConfig) (*ChartOperator, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	c := &ChartOperator{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
		tenant:    config.Tenant,
	}

	return c, nil
}

// Collect is the main metrics collection function.
func (c *ChartOperator) Collect(ch chan<- prometheus.Metric) error {
	clusters, err := c.helper.getTenantClusters()
	if err != nil {
		return microerror.Mask(err)
	}

	var g errgroup.Group

	for _, cluster := range clusters {
		g.Go(func() error {
			err := c.collectForTenantCluster(ch, cluster)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// Describe emits the description for the metrics collected here.
func (c *ChartOperator) Describe(ch chan<- *prometheus.Desc) error {
	// TODO
	return nil
}

func (c *ChartOperator) collectForTenantCluster(ch chan<- prometheus.Metric, cluster tenantCluster) error {
	ctx := context.Background()

	helmClient, err := c.tenant.NewHelmClient(ctx, cluster.id, cluster.apiDomain)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = helmClient.GetReleaseHistory(ctx, key.ChartOperatorReleaseName())
	if helmclient.IsReleaseNotFound(err) {

		// Return early. We will retry on the next execution.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
