package collector

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

var (
	clusterNodePools *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemCluster, "nodepools"),
		"Number of Node Pools in a cluster as provided by the MachineDeployment CRs with given cluster ID.",
		[]string{
			"cluster_id",
		},
		nil,
	)
)

type NodePoolConfig struct {
	CMAClient clientset.Interface
	Logger    micrologger.Logger
}

type NodePool struct {
	cmaClient clientset.Interface
	logger    micrologger.Logger
}

func NewNodePool(config NodePoolConfig) (*NodePool, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	np := &NodePool{
		cmaClient: config.CMAClient,
		logger:    config.Logger,
	}

	return np, nil
}

func (np *NodePool) Collect(ch chan<- prometheus.Metric) error {
	list, err := np.cmaClient.ClusterV1alpha1().Clusters(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, cluster := range list.Items {
		{
			clusterID := cluster.GetLabels()[label.Cluster]
			l := metav1.AddLabelToSelector(
				&metav1.LabelSelector{},
				label.Cluster,
				clusterID,
			)
			o := metav1.ListOptions{
				LabelSelector: labels.Set(l.MatchLabels).String(),
			}
			machineDeployments, err := np.cmaClient.ClusterV1alpha1().MachineDeployments(cluster.Namespace).List(o)
			if err != nil {
				return microerror.Mask(err)
			}

			ch <- prometheus.MustNewConstMetric(
				clusterNodePools,
				prometheus.GaugeValue,
				float64(len(machineDeployments.Items)),
				clusterID,
			)
		}
	}

	return nil
}

func (np *NodePool) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterNodePools
	return nil
}
