package collector

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

var (
	nodePools *prometheus.Desc = prometheus.NewDesc(
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

	list := &apiv1alpha2.MachineDeploymentList{}
	{
		np.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		err := np.k8sClient.CtrlClient().List(ctx, list)
		if err != nil {
			return microerror.Mask(err)
		}

		np.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d MachineDeployments for tenant cluster", len(list.Items)))
	}

	type nodes struct {
		nodePoolID string
		desired    int
		ready      int
	}

	clusterNodePools := make(map[string][]nodes)

	for _, md := range list.Items {
		clusterID := md.GetLabels()[label.Cluster]

		n := nodes{
			nodePoolID: md.GetLabels()[label.MachineDeployment],
			desired:    int(md.Status.Replicas),
			ready:      int(md.Status.ReadyReplicas),
		}

		clusterNodePools[clusterID] = append(clusterNodePools[clusterID], n)
	}

	for clusterID, nps := range clusterNodePools {
		{
			ch <- prometheus.MustNewConstMetric(
				nodePools,
				prometheus.GaugeValue,
				float64(len(nps)),
				clusterID,
			)
		}

		for _, n := range nps {
			{
				ch <- prometheus.MustNewConstMetric(
					nodePoolDesiredWorkers,
					prometheus.GaugeValue,
					float64(n.desired),
					clusterID,
					n.nodePoolID,
				)

				ch <- prometheus.MustNewConstMetric(
					nodePoolReadyWorkers,
					prometheus.GaugeValue,
					float64(n.ready),
					clusterID,
					n.nodePoolID,
				)
			}
		}
	}

	return nil
}

func (np *NodePool) Describe(ch chan<- *prometheus.Desc) error {
	ch <- nodePools
	ch <- nodePoolDesiredWorkers
	ch <- nodePoolReadyWorkers

	return nil
}
