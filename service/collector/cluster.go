package collector

import (
	"context"
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	apiextensionsconditions "github.com/giantswarm/apiextensions/v6/pkg/conditions"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v4/pkg/label"
	"github.com/giantswarm/cluster-operator/v4/pkg/project"
	"github.com/giantswarm/cluster-operator/v4/service/controller/key"
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

	NewCommonClusterObjectFunc func() infrastructurev1alpha3.CommonClusterObject
	Provider                   string
}

type Cluster struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	newCommonClusterObjectFunc func() infrastructurev1alpha3.CommonClusterObject
	provider                   string
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
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	c := &Cluster{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		newCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
		provider:                   config.Provider,
	}

	return c, nil
}

func (c *Cluster) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	var list apiv1beta1.ClusterList
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
		switch c.provider {
		case label.ProviderAWS:
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
					boolToFloat64(latest == infrastructurev1alpha3.ClusterStatusConditionCreating),
					key.ClusterID(&cl),
					key.ReleaseVersion(&cl),
					infrastructurev1alpha3.ClusterStatusConditionCreating,
				)
				ch <- prometheus.MustNewConstMetric(
					clusterStatus,
					prometheus.GaugeValue,
					boolToFloat64(latest == infrastructurev1alpha3.ClusterStatusConditionCreated),
					key.ClusterID(&cl),
					key.ReleaseVersion(&cl),
					infrastructurev1alpha3.ClusterStatusConditionCreated,
				)
				ch <- prometheus.MustNewConstMetric(
					clusterStatus,
					prometheus.GaugeValue,
					boolToFloat64(latest == infrastructurev1alpha3.ClusterStatusConditionUpdating),
					key.ClusterID(&cl),
					key.ReleaseVersion(&cl),
					infrastructurev1alpha3.ClusterStatusConditionUpdating,
				)
				ch <- prometheus.MustNewConstMetric(
					clusterStatus,
					prometheus.GaugeValue,
					boolToFloat64(latest == infrastructurev1alpha3.ClusterStatusConditionUpdated),
					key.ClusterID(&cl),
					key.ReleaseVersion(&cl),
					infrastructurev1alpha3.ClusterStatusConditionUpdated,
				)
				ch <- prometheus.MustNewConstMetric(
					clusterStatus,
					prometheus.GaugeValue,
					boolToFloat64(latest == infrastructurev1alpha3.ClusterStatusConditionDeleting),
					key.ClusterID(&cl),
					key.ReleaseVersion(&cl),
					infrastructurev1alpha3.ClusterStatusConditionDeleting,
				)
			}
		case label.ProviderAzure:
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(conditions.IsTrue(&cl, apiextensionsconditions.CreatingCondition)),
				key.ClusterID(&cl),
				key.ReleaseVersion(&cl),
				infrastructurev1alpha3.ClusterStatusConditionCreating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(conditions.IsTrue(&cl, apiv1beta1.ReadyCondition)),
				key.ClusterID(&cl),
				key.ReleaseVersion(&cl),
				infrastructurev1alpha3.ClusterStatusConditionCreated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatus,
				prometheus.GaugeValue,
				boolToFloat64(conditions.IsTrue(&cl, apiextensionsconditions.UpgradingCondition)),
				key.ClusterID(&cl),
				key.ReleaseVersion(&cl),
				infrastructurev1alpha3.ClusterStatusConditionUpdating,
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
