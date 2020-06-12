package collector

import (
	"github.com/giantswarm/exporterkit/histogramvec"
	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"
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
	clusterTransitionCreateHistogramVec *histogramvec.HistogramVec
	clusterTransitionUpdateHistogramVec *histogramvec.HistogramVec
	clusterTransitionDeleteHistogramVec *histogramvec.HistogramVec
}

//NewClusterTransition initiates cluster transition metrics
func NewClusterTransition() (*ClusterTransition, error) {
	var clusterTransitionCreateHistogramVec *histogramvec.HistogramVec
	var err error
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

	collector := &ClusterTransition{
		clusterTransitionCreateHistogramVec: clusterTransitionCreateHistogramVec,
		clusterTransitionUpdateHistogramVec: clusterTransitionUpdateHistogramVec,
		clusterTransitionDeleteHistogramVec: clusterTransitionDeleteHistogramVec,
	}
	return collector, nil
}

func (ct *ClusterTransition) Collect(ch chan<- prometheus.Metric) error {
	// ct.clusterTransitionCreateHistogramVec.Add(clusterID, observedTime.Seconds())
	// creation timestamp of metadata cluster?
	// Status Control Plane Initialized: false Infrastructure Ready: false

	//ct.clusterTransitionCreateHistogramVec.Ensure(clusters)
	//for host, histogram := range ct.clusterTransitionCreateHistogramVec.Histograms() {
	//ch <- prometheus.MustNewConstHistogram(
	//	clusterTransitionCreateDesc,
	//	histogram.Count(), histogram.Sum(), histogram.Buckets(),
	//	clusterID,
	//)
	return nil
}

func (ct *ClusterTransition) Describe(ch chan<- *prometheus.Desc) error {
	ch <- clusterTransitionCreateDesc
	ch <- clusterTransitionUpdateDesc
	ch <- clusterTransitionDeleteDesc

	return nil
}
