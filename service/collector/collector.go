package collector

import (
	"context"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	CertSearcher certs.Interface
	G8sClient    versioned.Interface
	Logger       micrologger.Logger
}

type Collector struct {
	g8sClient     versioned.Interface
	logger        micrologger.Logger
	tenantCluster tenantcluster.Interface
}

func New(config Config) (*Collector, error) {
	if config.CertSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertSearcher must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var err error

	var tenantCluster tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: config.CertSearcher,
			Logger:        config.Logger,

			CertID: certs.ClusterOperatorAPICert,
		}

		tenantCluster, err = tenantcluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Collector{
		g8sClient:     config.G8sClient,
		logger:        config.Logger,
		tenantCluster: tenantCluster,
	}

	return c, nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	c.logger.LogCtx(ctx, "level", "debug", "message", "collecting metrics")

	collectFuncs := []func(context.Context, chan<- prometheus.Metric){}

	var wg sync.WaitGroup

	for _, collectFunc := range collectFuncs {
		wg.Add(1)

		go func(collectFunc func(ctx context.Context, ch chan<- prometheus.Metric)) {
			defer wg.Done()
			collectFunc(ctx, ch)
		}(collectFunc)
	}

	wg.Wait()

	c.logger.LogCtx(ctx, "level", "debug", "message", "finished collecting metrics")
}
