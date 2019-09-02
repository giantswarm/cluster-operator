package collector

import (
	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/key"
)

const (
	subsystemCluster = "cluster"
)

var (
	clusterDescription *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemCluster, "status"),
		"Latest cluster status conditions as provided by the Cluster CR status.",
		[]string{
			"cluster_id",
			"status",
		},
		nil,
	)
)

type ClusterConfig struct {
	CMAClient clientset.Interface
	Logger    micrologger.Logger
}

type Cluster struct {
	cmaClient clientset.Interface
	logger    micrologger.Logger
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	c := &Cluster{
		cmaClient: config.CMAClient,
		logger:    config.Logger,
	}

	return c, nil
}

func (c *Cluster) Collect(ch chan<- prometheus.Metric) error {
	list, err := c.cmaClient.ClusterV1alpha1().Clusters(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, cluster := range list.Items {
		{
			latest := key.ClusterCommonStatus(cluster).LatestCondition()

			ch <- prometheus.MustNewConstMetric(
				clusterDescription,
				prometheus.GaugeValue,
				boolToFloat64(latest == v1alpha1.ClusterStatusConditionCreating),
				key.ClusterID(&cluster),
				v1alpha1.ClusterStatusConditionCreating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterDescription,
				prometheus.GaugeValue,
				boolToFloat64(latest == v1alpha1.ClusterStatusConditionCreated),
				key.ClusterID(&cluster),
				v1alpha1.ClusterStatusConditionCreated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterDescription,
				prometheus.GaugeValue,
				boolToFloat64(latest == v1alpha1.ClusterStatusConditionUpdating),
				key.ClusterID(&cluster),
				v1alpha1.ClusterStatusConditionUpdating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterDescription,
				prometheus.GaugeValue,
				boolToFloat64(latest == v1alpha1.ClusterStatusConditionUpdated),
				key.ClusterID(&cluster),
				v1alpha1.ClusterStatusConditionUpdated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterDescription,
				prometheus.GaugeValue,
				boolToFloat64(latest == v1alpha1.ClusterStatusConditionDeleting),
				key.ClusterID(&cluster),
				v1alpha1.ClusterStatusConditionDeleting,
			)
		}
	}

	return nil
}

func (c *Cluster) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterDescription
	return nil
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}

	return 0
}
