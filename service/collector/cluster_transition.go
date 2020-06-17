package collector

import (
	"context"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	createTransitionBuckets                      = []float64{600, 750, 900, 1050, 1200, 1350, 1500, 1650, 1800}
	updateTransitionBuckets                      = []float64{3600, 3900, 4200, 4500, 4800, 5100, 5400, 5700, 6000, 6300, 6600, 6900, 7200}
	deleteTransitionBuckets                      = []float64{3600, 3900, 4200, 4500, 4800, 5100, 5400, 5700, 6000, 6300, 6600, 6900, 7200}
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
	clusterTransitionDeleteDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemCluster, "delete_transition"),
		"Latest cluster deletion transition.",
		[]string{
			"cluster_id",
			"release_version",
		},
		nil,
	)
)

//ClusterTransition implements the ClusterTransition interface, exposing cluster transition information.
type ClusterTransition struct {
	clusterTransitionCreateHistogramVec *prometheus.HistogramVec
	clusterTransitionUpdateHistogramVec *prometheus.HistogramVec
	clusterTransitionDeleteHistogramVec *prometheus.HistogramVec

	k8sClient                  k8sclient.Interface
	logger                     micrologger.Logger
	newCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
}

type ClusterTransitionConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	NewCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
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

	var clusterTransitionCreateHistogramVec *prometheus.HistogramVec
	var labels = []string{"cluster_id", "release_version"}
	{
		c := prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystemCluster,
			Name:      "create_transition",
			Help:      "Latest cluster creation transition.",

			Buckets: createTransitionBuckets,
		}

		clusterTransitionCreateHistogramVec = prometheus.NewHistogramVec(c, labels)
	}
	var clusterTransitionUpdateHistogramVec *prometheus.HistogramVec
	{
		c := prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystemCluster,
			Name:      "update_transition",
			Help:      "Latest cluster update transition.",

			Buckets: updateTransitionBuckets,
		}

		clusterTransitionUpdateHistogramVec = prometheus.NewHistogramVec(c, labels)
	}
	var clusterTransitionDeleteHistogramVec *prometheus.HistogramVec
	{
		c := prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystemCluster,
			Name:      "delete_transition",
			Help:      "Latest cluster deletion transition.",
			Buckets:   deleteTransitionBuckets,
		}

		clusterTransitionDeleteHistogramVec = prometheus.NewHistogramVec(c, labels)
	}

	collector := &ClusterTransition{
		clusterTransitionCreateHistogramVec: clusterTransitionCreateHistogramVec,
		clusterTransitionUpdateHistogramVec: clusterTransitionUpdateHistogramVec,
		clusterTransitionDeleteHistogramVec: clusterTransitionDeleteHistogramVec,
	}
	return collector, nil
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
			if err != nil {
				return microerror.Mask(err)
			}
		}

		now := time.Now()
		{
			if cr.GetCommonClusterStatus().HasCreatingCondition() && !cr.GetCommonClusterStatus().HasCreatedCondition() {
				clusterTransistionCreating := cr.GetCommonClusterStatus().GetCreatingCondition().LastTransitionTime
				maxCreateInterval := clusterTransistionCreating.Add(30 * time.Minute)

				if now.After(maxCreateInterval) {
					ct.clusterTransitionCreateHistogramVec.WithLabelValues(cr.GetClusterName()).Observe(float64(999999999999))
				}

			}

			if cr.GetCommonClusterStatus().HasUpdatingCondition() && !cr.GetCommonClusterStatus().HasUpdatedCondition() {
				clusterTransistionCreating := cr.GetCommonClusterStatus().GetUpdatingCondition().LastTransitionTime
				maxUpdateInterval := clusterTransistionCreating.Add(2 * time.Hour)

				if now.After(maxUpdateInterval) {
					ct.clusterTransitionCreateHistogramVec.WithLabelValues(cr.GetClusterName()).Observe(float64(999999999999))
				}

			}

			if cr.GetCommonClusterStatus().HasCreatingCondition() && cr.GetCommonClusterStatus().HasCreatedCondition() {
				clusterTransistionCreated := cr.GetCommonClusterStatus().GetCreatedCondition().LastTransitionTime.Unix()
				clusterTransitionCreating := cr.GetCommonClusterStatus().GetCreatingCondition().LastTransitionTime.Unix()

				deltaCreate := clusterTransistionCreated - clusterTransitionCreating
				ct.clusterTransitionCreateHistogramVec.WithLabelValues(cr.GetClusterName()).Observe(float64(deltaCreate))
			}

			if cr.GetCommonClusterStatus().HasUpdatingCondition() && cr.GetCommonClusterStatus().HasUpdatedCondition() {
				clusterTransitionUpdating := cr.GetCommonClusterStatus().GetUpdatedCondition().LastTransitionTime.Unix()
				clusterTransitionUpdated := cr.GetCommonClusterStatus().GetUpdatingCondition().LastTransitionTime.Unix()

				deltaUpdate := clusterTransitionUpdated - clusterTransitionUpdating
				ct.clusterTransitionCreateHistogramVec.WithLabelValues(cr.GetClusterName()).Observe(float64(deltaUpdate))
			}

			//deleting figure out howto get deleted

		}

	}
	return nil
}

func (ct *ClusterTransition) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterTransitionCreateDesc
	ch <- clusterTransitionUpdateDesc
	ch <- clusterTransitionDeleteDesc

	return nil
}
