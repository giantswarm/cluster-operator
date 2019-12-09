package collector

import (
	"github.com/giantswarm/certs"
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type SetConfig struct {
	CertSearcher certs.Interface
	K8sClient    k8sclient.Interface
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

	var nodePoolCollector *NodePool
	{
		c := NodePoolConfig{
			CMAClient: config.CMAClient,
			Logger:    config.Logger,
		}

		nodePoolCollector, err = NewNodePool(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: []collector.Interface{
				clusterCollector,
				nodePoolCollector,
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
