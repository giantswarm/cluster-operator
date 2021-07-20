package collector

import (
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type SetConfig struct {
	CertSearcher certs.Interface
	K8sClient    k8sclient.Interface
	Logger       micrologger.Logger

	NewCommonClusterObjectFunc func() infrastructurev1alpha3.CommonClusterObject
}

// Set is basically only a wrapper for the operator's collector implementations.
// It eases the initialization and prevents some weird import mess so we do not
// have to alias packages.
type Set struct {
	*collector.Set
}

func NewSet(config SetConfig) (*Set, error) {
	var err error

	var clusterCollector *Cluster
	{
		c := ClusterConfig{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
		}

		clusterCollector, err = NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var nodePoolCollector *NodePool
	{
		c := NodePoolConfig{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		nodePoolCollector, err = NewNodePool(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterTransitionCollector *ClusterTransition
	{
		c := ClusterTransitionConfig{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
		}

		clusterTransitionCollector, err = NewClusterTransition(c)
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
				clusterTransitionCollector,
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
