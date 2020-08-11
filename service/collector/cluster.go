package collector

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

var (
	clusterStatus *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemCluster, "status"),
		"Latest cluster status conditions as provided by the Cluster CR status.",
		[]string{
			"cluster_id",
			"release_version",
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

	var list apiv1alpha2.ClusterList
	{
		err := c.k8sClient.CtrlClient().List(
			ctx,
			&list,
			client.MatchingLabels{label.OperatorVersion: project.Version()},
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	for _, cl := range list.Items {
		cl := cl // dereferencing pointer value into new scope

		cr := c.newCommonClusterObjectFunc()
		{
			err := c.k8sClient.CtrlClient().Get(
				ctx,
				key.ObjRefToNamespacedName(key.ObjRefFromCluster(cl)),
				cr,
			)
			if apierrors.IsNotFound(err) {
				c.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("could not find object reference %#q", cl.GetName()))
				continue
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		{
			latest := cr.GetCommonClusterStatus().LatestCondition()

			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionCreating),
				key.ClusterID(&cl),
				key.ReleaseVersion(&cl),
				infrastructurev1alpha2.ClusterStatusConditionCreating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionCreated),
				key.ClusterID(&cl),
				key.ReleaseVersion(&cl),
				infrastructurev1alpha2.ClusterStatusConditionCreated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionUpdating),
				key.ClusterID(&cl),
				key.ReleaseVersion(&cl),
				infrastructurev1alpha2.ClusterStatusConditionUpdating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionUpdated),
				key.ClusterID(&cl),
				key.ReleaseVersion(&cl),
				infrastructurev1alpha2.ClusterStatusConditionUpdated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(latest == infrastructurev1alpha2.ClusterStatusConditionDeleting),
				key.ClusterID(&cl),
				key.ReleaseVersion(&cl),
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
