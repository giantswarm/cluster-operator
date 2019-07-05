package collector

import (
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	subsystemCluster = "cluster"
)

var (
	clusterStatusCollectorDescription *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemCluster, "status"),
		"Latest cluster status conditions as provided by the Cluster CR status.",
		[]string{
			"cluster_id",
			"status",
		},
		nil,
	)
)

type ClusterStatusCollectorConfig struct {
	CMAClient clientset.Interface
	Logger    micrologger.Logger
}

type ClusterStatusCollector struct {
	cmaClient clientset.Interface
	logger    micrologger.Logger
}

func NewClusterStatusCollector(config ClusterStatusCollectorConfig) (*ClusterStatusCollector, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	c := &ClusterStatusCollector{
		cmaClient: config.CMAClient,
		logger:    config.Logger,
	}

	return c, nil
}

func (c *ClusterStatusCollector) Collect(ch chan<- prometheus.Metric) error {
	list, err := c.cmaClient.ClusterV1alpha1().Clusters(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, cluster := range list.Items {
		{
		condition := key.ClusterCommonStatus(cluster).Conditions

			ch <- prometheus.MustNewConstMetric(
				clusterStatusCollectorDescription,
				prometheus.GaugeValue,
				float64(boolToInt(p.ClusterStatus().HasCreatingCondition())),
				m.GetName(),
				providerv1alpha1.StatusClusterTypeCreating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatusCollectorDescription,
				prometheus.GaugeValue,
				float64(boolToInt(p.ClusterStatus().HasCreatedCondition())),
				m.GetName(),
				providerv1alpha1.StatusClusterTypeCreated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatusCollectorDescription,
				prometheus.GaugeValue,
				float64(boolToInt(p.ClusterStatus().HasUpdatingCondition())),
				m.GetName(),
				providerv1alpha1.StatusClusterTypeUpdating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatusCollectorDescription,
				prometheus.GaugeValue,
				float64(boolToInt(p.ClusterStatus().HasUpdatedCondition())),
				m.GetName(),
				providerv1alpha1.StatusClusterTypeUpdated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatusCollectorDescription,
				prometheus.GaugeValue,
				float64(boolToInt(p.ClusterStatus().HasDeletingCondition())),
				m.GetName(),
				providerv1alpha1.StatusClusterTypeDeleting,
			)
		}
	}
}

func (c *ClusterStatusCollector) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterStatusCollectorDescription
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}
