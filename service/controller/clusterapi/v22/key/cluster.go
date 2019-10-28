package key

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func AWSClusterConfigName(cluster v1alpha1.Cluster) string {
	return cluster.Name
}

func ClusterAPIEndpoint(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("api.%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
}

func ClusterBaseDomain(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Cluster.DNS.Domain
}

func ClusterCredentialSecretName(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.CredentialSecret.Name
}

func ClusterCredentialSecretNamespace(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.CredentialSecret.Namespace
}

func ClusterDNSZone(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s.k8s.%s", ClusterID(&cluster), clusterProviderSpec(cluster).Cluster.DNS.Domain)
}

func ClusterMasterAZ(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.Master.AvailabilityZone
}

func ClusterMasterInstanceType(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.Master.InstanceType
}

func ClusterName(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).ClusterName
}

func IsProviderSpecForAWS(cluster v1alpha1.Cluster) bool {
	_, err := g8sClusterSpecFromCMAClusterSpec(cluster.Spec.ProviderSpec)
	return err == nil
}

func TenantBaseDomain(cluster v1alpha1.Cluster) string {
	return fmt.Sprintf("%s.k8s.%s", ClusterID(&cluster), ClusterBaseDomain(cluster))
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
