package collector

import (
	"context"

	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

var (
	nodePoolCount *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemNodePool, "count"),
		"Number of Node Pools in a cluster as provided by the MachineDeployment CRs associated with a given cluster ID.",
		[]string{
			"cluster_id",
		},
		nil,
	)

	nodePoolDesiredWorkers *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemNodePool, "desired_workers"),
		"Number of desired workers in all node pools for a specific cluster as provided by the MachineDeployment CRs associated with a given cluster ID.",
		[]string{
			"cluster_id",
			"node_pool_id",
		},
		nil,
	)

	nodePoolReadyWorkers *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemNodePool, "ready_workers"),
		"Number of ready workers in all node pools for a specific cluster as provided by the MachineDeployment CRs associated with a given cluster ID.",
		[]string{
			"cluster_id",
			"node_pool_id",
		},
		nil,
	)
)

type NodePoolConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type NodePool struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func NewNodePool(config NodePoolConfig) (*NodePool, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	np := &NodePool{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return np, nil
}

func (np *NodePool) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	var list apiv1alpha2.MachineDeploymentList
	{
		err := np.k8sClient.CtrlClient().List(
			ctx,
			&list,
			client.MatchingLabels{label.OperatorVersion: project.Version()},
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	type nodePool struct {
		id      string
		desired int
		ready   int
	}

	nodePoolMap := make(map[string][]nodePool)

	for _, md := range list.Items {
		np := nodePool{
			id:      key.MachineDeployment(&md),
			desired: int(md.Status.Replicas),
			ready:   int(md.Status.ReadyReplicas),
		}

		nodePoolMap[key.ClusterID(&md)] = append(nodePoolMap[key.ClusterID(&md)], np)
	}

	for cid, nps := range nodePoolMap {
		{
			ch <- prometheus.MustNewConstMetric(
				nodePoolCount,
				prometheus.GaugeValue,
				float64(len(nps)),
				cid,
			)
		}

		for _, np := range nps {
			ch <- prometheus.MustNewConstMetric(
				nodePoolDesiredWorkers,
				prometheus.GaugeValue,
				float64(np.desired),
				cid,
				np.id,
			)

			ch <- prometheus.MustNewConstMetric(
				nodePoolReadyWorkers,
				prometheus.GaugeValue,
				float64(np.ready),
				cid,
				np.id,
			)
		}
	}

	return nil
}

func (np *NodePool) Describe(ch chan<- *prometheus.Desc) error {
	ch <- nodePoolCount
	ch <- nodePoolDesiredWorkers
	ch <- nodePoolReadyWorkers

	return nil
}
