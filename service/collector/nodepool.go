package collector

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

var (
	nodePools *prometheus.Desc = prometheus.NewDesc(
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
	list, err := np.cmaClient.ClusterV1alpha1().MachineDeployments(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	clusterNodePools := make(map[string]int)

	for _, md := range list.Items {
		clusterID := md.GetLabels()[label.Cluster]

		clusterNodePools[clusterID] = clusterNodePools[clusterID] + 1
	}

	for clusterID, nodePoolCount := range clusterNodePools {
		{
			ch <- prometheus.MustNewConstMetric(
				nodePools,
				prometheus.GaugeValue,
				float64(nodePoolCount),
				clusterID,
			)
		}
	}

	return nil
}

func (np *NodePool) Describe(ch chan<- *prometheus.Desc) error {
	ch <- nodePools
	return nil
}
