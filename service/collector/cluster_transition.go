package collector

import (
	"context"
	"fmt"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
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
	clusterTransitionCreateDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemCluster, "create_transition"),
		"Latest cluster creation transition.",
		[]string{
			"cluster_id",
			"release_version",
		},
		nil,
	)
	clusterTransitionUpdateDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemCluster, "update_transition"),
		"Latest cluster update transition.",
		[]string{
			"cluster_id",
			"release_version",
		},
		nil,
	)
)

type ClusterTransitionConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	NewCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
}

// ClusterTransition implements the ClusterTransition interface, exposing
// cluster transition information.
type ClusterTransition struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	newCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
}

//NewClusterTransition initiates cluster transition metrics
func NewClusterTransition(config ClusterTransitionConfig) (*ClusterTransition, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.NewCommonClusterObjectFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.NewCommonClusterObjectFunc must not be empty", config)
	}

	ct := &ClusterTransition{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		newCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
	}

	return ct, nil
}

func (ct *ClusterTransition) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	var list apiv1alpha2.ClusterList
	{
		err := ct.k8sClient.CtrlClient().List(
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

		cr := ct.newCommonClusterObjectFunc()
		{
			err := ct.k8sClient.CtrlClient().Get(
				ctx,
				key.ObjRefToNamespacedName(key.ObjRefFromCluster(cl)),
				cr,
			)
			if apierrors.IsNotFound(err) {
				ct.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("could not find object reference %#q", cl.GetName()))
				continue
			} else if err != nil {
				return microerror.Mask(err)
			}
		}
		{
			created, createTime := getCreateMetrics(cr.GetCommonClusterStatus())
			if created {
				ch <- prometheus.MustNewConstMetric(
					clusterTransitionCreateDesc,
					prometheus.GaugeValue,
					createTime,
					key.ClusterID(cr),
					key.ReleaseVersion(cr),
				)
			}
			updated, updateTime := getUpdateMetrics(cr.GetCommonClusterStatus())
			if updated {
				ch <- prometheus.MustNewConstMetric(
					clusterTransitionUpdateDesc,
					prometheus.GaugeValue,
					updateTime,
					key.ClusterID(cr),
					key.ReleaseVersion(cr),
				)
			}
		}
	}
	return nil
}

func getCreateMetrics(status infrastructurev1alpha2.CommonClusterStatus) (bool, float64) {
	if status.HasCreatingCondition() && status.HasCreatedCondition() {
		t1 := status.GetCreatingCondition().LastTransitionTime.Time
		t2 := status.GetCreatedCondition().LastTransitionTime.Time
		return true, t2.Sub(t1).Seconds()
	}

	if status.HasCreatingCondition() && !status.HasCreatedCondition() {
		t1 := status.GetCreatingCondition().LastTransitionTime.Time

		// If the Creating condition is too old without having any
		// Created condition given, we put the cluster into the last
		// bucket and consider it invalid in that regard.
		if time.Now().After(t1.Add(30 * time.Minute)) {
			return true, float64(999999999999)
		}
	}
	return false, 0
}
func getUpdateMetrics(status infrastructurev1alpha2.CommonClusterStatus) (bool, float64) {

	if status.HasUpdatingCondition() && status.HasUpdatedCondition() {
		t1 := status.GetUpdatingCondition().LastTransitionTime.Time
		t2 := status.GetUpdatedCondition().LastTransitionTime.Time
		if t2.Sub(t1).Seconds() > 0 {
			return true, t2.Sub(t1).Seconds()
		}
	}

	if status.HasUpdatingCondition() {
		t1 := status.GetUpdatingCondition().LastTransitionTime.Time

		// If the Updating condition is too old without having any
		// Updated condition given, we put the cluster into the last
		// bucket and consider it invalid in that regard.
		if time.Now().After(t1.Add(2 * time.Hour)) {
			return true, float64(999999999999)
		}
	}
	return false, 0
}

func (ct *ClusterTransition) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterTransitionCreateDesc
	ch <- clusterTransitionUpdateDesc

	return nil
}
