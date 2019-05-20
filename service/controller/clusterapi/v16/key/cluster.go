package key

import (
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func ClusterID(cluster v1alpha1.Cluster) string {
	return clusterProviderStatus(cluster).Cluster.ID
}

func AWSClusterConfigName(cluster v1alpha1.Cluster) string {
	return clusterProviderStatus(cluster).Cluster.ID
}

func IsProviderSpecForAWS(cluster v1alpha1.Cluster) bool {
	_, err := g8sClusterSpecFromCMAClusterSpec(cluster.Spec.ProviderSpec)
	return err == nil
}

func IsProviderStatusForAWS(cluster v1alpha1.Cluster) bool {
	_, err := g8sClusterStatusFromCMAClusterStatus(cluster.Status.ProviderStatus)
	return err == nil
}

func ToCluster(v interface{}) (v1alpha1.Cluster, error) {
	if v == nil {
		return v1alpha1.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.Cluster{}, v)
	}

	p, ok := v.(*v1alpha1.Cluster)
	if !ok {
		return v1alpha1.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.Cluster{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
