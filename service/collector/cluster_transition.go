package collector

import (
	"context"
	"fmt"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/exporterkit/histogramvec"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

var (
	// createTransitionBuckets ranges from 10 minutes to 30 minutes.
	createTransitionBuckets = []float64{600, 750, 900, 1050, 1200, 1350, 1500, 1650, 1800}
	// updateTransitionBuckets ranges from 1 hour to 2 hours.
	updateTransitionBuckets = []float64{3600, 3900, 4200, 4500, 4800, 5100, 5400, 5700, 6000, 6300, 6600, 6900, 7200}
	// deleteTransitionBuckets ranges from 1 hour to 2 hours.
	deleteTransitionBuckets = []float64{3600, 3900, 4200, 4500, 4800, 5100, 5400, 5700, 6000, 6300, 6600, 6900, 7200}
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
	//clusterTransitionDeleteDesc *prometheus.Desc = prometheus.NewDesc(
	//	prometheus.BuildFQName(namespace, subsystemCluster, "delete_transition"),
	//	"Latest cluster deletion transition.",
	//	[]string{
	//		"cluster_id",
	//		"release_version",
	//	},
	//	nil,
	//)
)

type ClusterTransitionConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	NewCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
}

// ClusterTransition implements the ClusterTransition interface, exposing
// cluster transition information.
type ClusterTransition struct {
	k8sClient                  k8sclient.Interface
	logger                     micrologger.Logger
	newCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject

	clusterTransitionCreateHistogramVec *histogramvec.HistogramVec
	clusterTransitionUpdateHistogramVec *histogramvec.HistogramVec
	clusterTransitionDeleteHistogramVec *histogramvec.HistogramVec
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

	var err error

	var clusterTransitionCreateHistogramVec *histogramvec.HistogramVec
	{
		c := histogramvec.Config{
			BucketLimits: createTransitionBuckets,
		}

		clusterTransitionCreateHistogramVec, err = histogramvec.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterTransitionUpdateHistogramVec *histogramvec.HistogramVec
	{
		c := histogramvec.Config{
			BucketLimits: updateTransitionBuckets,
		}

		clusterTransitionUpdateHistogramVec, err = histogramvec.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterTransitionDeleteHistogramVec *histogramvec.HistogramVec
	{
		c := histogramvec.Config{
			BucketLimits: deleteTransitionBuckets,
		}

		clusterTransitionDeleteHistogramVec, err = histogramvec.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	ct := &ClusterTransition{
		k8sClient:                  config.K8sClient,
		logger:                     config.Logger,
		newCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,

		clusterTransitionCreateHistogramVec: clusterTransitionCreateHistogramVec,
		clusterTransitionUpdateHistogramVec: clusterTransitionUpdateHistogramVec,
		clusterTransitionDeleteHistogramVec: clusterTransitionDeleteHistogramVec,
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
			//client.MatchingLabels{label.OperatorVersion: project.Version()},
			client.MatchingLabels{label.OperatorVersion: "2.2.0"},
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var clusters []string
	releases := map[string]string{}
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

		clusters = append(clusters, cr.GetName())
		releases[cr.GetName()] = key.ReleaseVersion(cr)
		var err error
		{
			{
				clusterHistogram := ct.clusterTransitionCreateHistogramVec.Histograms()
				_, ok := clusterHistogram[cr.GetName()]

				if cr.GetCommonClusterStatus().HasCreatingCondition() && cr.GetCommonClusterStatus().HasCreatedCondition() && !ok {
					t1 := cr.GetCommonClusterStatus().GetCreatingCondition().LastTransitionTime.Time
					t2 := cr.GetCommonClusterStatus().GetCreatedCondition().LastTransitionTime.Time
					err = ct.clusterTransitionCreateHistogramVec.Add(cr.GetName(), t2.Sub(t1).Seconds())
					if err != nil {
						return microerror.Mask(err)
					}
				}

				if cr.GetCommonClusterStatus().HasCreatingCondition() && !cr.GetCommonClusterStatus().HasCreatedCondition() {
					t1 := cr.GetCommonClusterStatus().GetCreatingCondition().LastTransitionTime.Time
					maxInterval := createTransitionBuckets[len(createTransitionBuckets)-1]

					// If the Creating condition is too old without having any
					// Created condition given, we put the cluster into the last
					// bucket and consider it invalid in that regard.
					if time.Now().After(t1.Add(time.Duration(maxInterval)*time.Minute)) && !ok {
						err = ct.clusterTransitionCreateHistogramVec.Add(cr.GetName(), float64(999999999999))
						if err != nil {
							return microerror.Mask(err)
						}
					}
				}
			}
			{
				clusterHistogram := ct.clusterTransitionUpdateHistogramVec.Histograms()
				_, ok := clusterHistogram[cr.GetName()]

				if cr.GetCommonClusterStatus().HasUpdatingCondition() && cr.GetCommonClusterStatus().HasUpdatedCondition() && !ok {
					t1 := cr.GetCommonClusterStatus().GetUpdatingCondition().LastTransitionTime.Time
					t2 := cr.GetCommonClusterStatus().GetUpdatedCondition().LastTransitionTime.Time
					err = ct.clusterTransitionUpdateHistogramVec.Add(cr.GetName(), t2.Sub(t1).Seconds())
					if err != nil {
						return microerror.Mask(err)
					}
				}

				if cr.GetCommonClusterStatus().HasUpdatingCondition() && !cr.GetCommonClusterStatus().HasUpdatedCondition() {
					t1 := cr.GetCommonClusterStatus().GetUpdatingCondition().LastTransitionTime.Time
					maxInterval := updateTransitionBuckets[len(updateTransitionBuckets)-1]

					// If the Updating condition is too old without having any
					// Updated condition given, we put the cluster into the last
					// bucket and consider it invalid in that regard.
					if time.Now().After(t1.Add(time.Duration(maxInterval)*time.Minute)) && !ok {
						err = ct.clusterTransitionUpdateHistogramVec.Add(cr.GetName(), float64(999999999999))
						if err != nil {
							return microerror.Mask(err)
						}
					}
				}
			}
		}
	}
	ct.clusterTransitionCreateHistogramVec.Ensure(clusters)

	for cluster, histogram := range ct.clusterTransitionCreateHistogramVec.Histograms() {
		ch <- prometheus.MustNewConstHistogram(
			clusterTransitionCreateDesc,
			histogram.Count(), histogram.Sum(), histogram.Buckets(),
			cluster,
			releases[cluster],
		)
	}

	return nil
}

func (ct *ClusterTransition) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterTransitionCreateDesc
	ch <- clusterTransitionUpdateDesc
	//	ch <- clusterTransitionDeleteDesc

	return nil
}
