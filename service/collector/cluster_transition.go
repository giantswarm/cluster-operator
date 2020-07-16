package collector

import (
	"context"
	"fmt"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/key"
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
		k8sClient:                  config.K8sClient,
		logger:                     config.Logger,
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
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		{
			{
				if cr.GetCommonClusterStatus().HasCreatingCondition() && cr.GetCommonClusterStatus().HasCreatedCondition() {
					t1 := cr.GetCommonClusterStatus().GetCreatingCondition().LastTransitionTime.Time
					t2 := cr.GetCommonClusterStatus().GetCreatedCondition().LastTransitionTime.Time
					ch <- prometheus.MustNewConstMetric(
						clusterTransitionCreateDesc,
						prometheus.GaugeValue,
						t2.Sub(t1).Seconds(),
						key.ClusterID(cr),
						key.ReleaseVersion(cr),
					)
				}

				if cr.GetCommonClusterStatus().HasCreatingCondition() && !cr.GetCommonClusterStatus().HasCreatedCondition() {
					t1 := cr.GetCommonClusterStatus().GetCreatingCondition().LastTransitionTime.Time

					// If the Creating condition is too old without having any
					// Created condition given, we put the cluster into the last
					// bucket and consider it invalid in that regard.
					if time.Now().After(t1.Add(30 * time.Minute)) {
						ch <- prometheus.MustNewConstMetric(
							clusterTransitionCreateDesc,
							prometheus.GaugeValue,
							float64(999999999999),
							key.ClusterID(cr),
							key.ReleaseVersion(cr),
						)
					}
				}
			}
			{
				if cr.GetCommonClusterStatus().HasUpdatingCondition() && cr.GetCommonClusterStatus().HasUpdatedCondition() {
					t1 := cr.GetCommonClusterStatus().GetUpdatingCondition().LastTransitionTime.Time
					t2 := cr.GetCommonClusterStatus().GetUpdatedCondition().LastTransitionTime.Time
					ch <- prometheus.MustNewConstMetric(
						clusterTransitionUpdateDesc,
						prometheus.GaugeValue,
						t2.Sub(t1).Seconds(),
						key.ClusterID(cr),
						key.ReleaseVersion(cr),
					)
				}

				if cr.GetCommonClusterStatus().HasUpdatingCondition() && !cr.GetCommonClusterStatus().HasUpdatedCondition() {
					t1 := cr.GetCommonClusterStatus().GetUpdatingCondition().LastTransitionTime.Time

					// If the Updating condition is too old without having any
					// Updated condition given, we put the cluster into the last
					// bucket and consider it invalid in that regard.
					if time.Now().After(t1.Add(2 * time.Hour)) {
						ch <- prometheus.MustNewConstMetric(
							clusterTransitionCreateDesc,
							prometheus.GaugeValue,
							float64(999999999999),
							key.ClusterID(cr),
							key.ReleaseVersion(cr),
						)
					}
				}
			}
		}
	}
	return nil
}

func (ct *ClusterTransition) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterTransitionCreateDesc
	ch <- clusterTransitionUpdateDesc

	return nil
}
