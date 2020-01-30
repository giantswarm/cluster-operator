package collector

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

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

	NewCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
}

type Cluster struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	newCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.NewCommonClusterObjectFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.NewCommonClusterObjectFunc must not be empty", config)
	}

	c := &Cluster{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		newCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
	}

	return c, nil
}

func (c *Cluster) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()
	list := &apiv1alpha3.ClusterList{}

	err := c.k8sClient.CtrlClient().List(ctx, list)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, cl := range list.Items {
		cr := c.newCommonClusterObjectFunc()
		err := c.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(key.ObjRefFromCluster(cl)), cr)
		if err != nil {
			return microerror.Mask(err)
		}

		{
			latest := cr.GetCommonClusterStatus().LatestCondition()

			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionCreating),
				key.ClusterID(&cl),
				infrastructurev1alpha2.ClusterStatusConditionCreating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionCreated),
				key.ClusterID(&cl),
				infrastructurev1alpha2.ClusterStatusConditionCreated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionUpdating),
				key.ClusterID(&cl),
				infrastructurev1alpha2.ClusterStatusConditionUpdating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionUpdated),
				key.ClusterID(&cl),
				infrastructurev1alpha2.ClusterStatusConditionUpdated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionDeleting),
				key.ClusterID(&cl),
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
