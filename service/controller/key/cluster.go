package key

import (
	"fmt"

	"github.com/giantswarm/microerror"
	clusterv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

func ClusterAPIEndpoint(cluster clusterv1alpha2.Cluster) string {
	return fmt.Sprintf("api.%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
}

func ClusterBaseDomain(cluster clusterv1alpha2.Cluster) string {
	return clusterProviderSpec(cluster).Cluster.DNS.Domain
}

func TenantBaseDomain(cluster clusterv1alpha2.Cluster) string {
	return fmt.Sprintf("%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
}

func ToCluster(v interface{}) (clusterv1alpha2.Cluster, error) {
	if v == nil {
		return clusterv1alpha2.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &clusterv1alpha2.Cluster{}, v)
	}

	p, ok := v.(*clusterv1alpha2.Cluster)
	if !ok {
		return clusterv1alpha2.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &clusterv1alpha2.Cluster{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
