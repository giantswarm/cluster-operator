package collector

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

type SetConfig struct {
	CertSearcher certs.Interface
	CMAClient    clientset.Interface
	G8sClient    versioned.Interface
	Logger       micrologger.Logger
}

// Set is basically only a wrapper for the operator's collector implementations.
// It eases the iniitialization and prevents some weird import mess so we do not
// have to alias packages.
type Set struct {
	*collector.Set
}

func NewSet(config SetConfig) (*Set, error) {
	var err error

	var helper *helper
	{
		c := helperConfig{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		helper, err = newHelper(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

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

	var chartOperatorCollector *ChartOperator
	{
		c := ChartOperatorConfig{
			G8sClient:     config.G8sClient,
			Helper:        helper,
			Logger:        config.Logger,
			TenantCluster: tenantCluster,
		}

		chartOperatorCollector, err = NewChartOperator(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterCollector *Cluster
	{
		c := ClusterConfig{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		clusterCollector, err = NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: []collector.Interface{
				chartOperatorCollector,
				clusterCollector,
			},
			Logger: config.Logger,
		}

		collectorSet, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Set{
		Set: collectorSet,
	}

	return s, nil
}
