package collector

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/service/controller/key"
)

var (
	clusterStatus *prometheus.Desc = prometheus.NewDesc(
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
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type Cluster struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	c := &Cluster{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return c, nil
}

func (c *Cluster) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	list := &apiv1alpha2.ClusterList{}

	err := c.k8sClient.CtrlClient().List(ctx, list)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, cluster := range list.Items {
		statusReader := &infrastructurev1alpha2.StatusReader{}
		err := c.k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: cluster.Spec.InfrastructureRef.Name, Namespace: cluster.Spec.InfrastructureRef.Namespace}, statusReader)
		if err != nil {
			return microerror.Mask(err)
		}

		{
			latest := statusReader.Status.Cluster.LatestCondition()

			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionCreating),
				key.ClusterID(&cluster),
				infrastructurev1alpha2.ClusterStatusConditionCreating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionCreated),
				key.ClusterID(&cluster),
				infrastructurev1alpha2.ClusterStatusConditionCreated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionUpdating),
				key.ClusterID(&cluster),
				infrastructurev1alpha2.ClusterStatusConditionUpdating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionUpdated),
				key.ClusterID(&cluster),
				infrastructurev1alpha2.ClusterStatusConditionUpdated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionDeleting),
				key.ClusterID(&cluster),
				infrastructurev1alpha2.ClusterStatusConditionDeleting,
			)
		}
	}

	return nil
}

func (c *Cluster) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterStatus
	return nil
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}

	return 0
}
