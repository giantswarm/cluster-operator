package collector

import (
	"github.com/giantswarm/exporterkit/histogramvec"
	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	createTransitionBucketStart      = 0.001
	createTransitionBucketFactor     = 2
	createTransitionBucketNumBuckets = 5
	updateTransitionBucketStart      = 0.001
	updateTransitionBucketFactor     = 2
	updateTransitionBucketNumBuckets = 5
	deleteTransitionBucketStart      = 0.001
	deleteTransitionBucketFactor     = 2
	deleteTransitionBucketNumBuckets = 5
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

// ClusterTransitionCollect implements the ClusterTransitionCollector interface, exposing cluster transition information.
type ClusterTransitionCollector struct {
	clusterTransitionCreateHistogramVec  *histogramvec.HistogramVec
	clusterTransitionUpdateHistogramVec  *histogramvec.HistogramVec
	clusterTransitionDeleteHistogramVec  *histogramvec.HistogramVec
	clusterTransitionCreateHistogramDesc *prometheus.Desc
	clusterTransitionUpdateHistogramDesc *prometheus.Desc
	clusterTransitionDeleteHistogramDesc *prometheus.Desc
}

//TODO
func NewClusterTransition() (*ClusterTransitionCollector, error) {
	var clusterTransitionCreateHistogramVec *histogramvec.HistogramVec
	var err error
	{
		c := histogramvec.Config{
			BucketLimits: prometheus.ExponentialBuckets(
				createTransitionBucketStart,
				createTransitionBucketFactor,
				createTransitionBucketNumBuckets,
			),
		}
		clusterTransitionCreateHistogramVec, err = histogramvec.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	var clusterTransitionUpdateHistogramVec *histogramvec.HistogramVec
	{
		c := histogramvec.Config{
			BucketLimits: prometheus.ExponentialBuckets(
				updateTransitionBucketStart,
				updateTransitionBucketFactor,
				updateTransitionBucketNumBuckets,
			),
		}
		clusterTransitionUpdateHistogramVec, err = histogramvec.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	var clusterTransitionDeleteHistogramVec *histogramvec.HistogramVec
	{
		c := histogramvec.Config{
			BucketLimits: prometheus.ExponentialBuckets(
				deleteTransitionBucketStart,
				deleteTransitionBucketFactor,
				deleteTransitionBucketNumBuckets),
		}
		clusterTransitionDeleteHistogramVec, err = histogramvec.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	collector := &ClusterTransitionCollector{
		clusterTransitionCreateHistogramVec:  clusterTransitionCreateHistogramVec,
		clusterTransitionCreateHistogramDesc: clusterTransitionCreateDesc,
		clusterTransitionUpdateHistogramVec:  clusterTransitionUpdateHistogramVec,
		clusterTransitionUpdateHistogramDesc: clusterTransitionUpdateDesc,
		clusterTransitionDeleteHistogramVec:  clusterTransitionDeleteHistogramVec,
		clusterTransitionDeleteHistogramDesc: clusterTransitionDeleteDesc,
	}
	return collector, nil
}
