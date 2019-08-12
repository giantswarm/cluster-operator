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

func ClusterAvailabilityZones(cluster v1alpha1.Cluster) []string {
	azMap := make(map[string]struct{})

	azMap[ClusterMasterAZ(cluster)] = struct{}{}

	// TODO: Extract AZs from MachineDeployments

	azs := make([]string, 0, len(azMap))
	for az := range azMap {
		azs = append(azs, az)
	}
	return azs
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

// EncryptionKeySecretName generates name for a Kubernetes secret based on
// information in given v1alpha1.ClusterGuestConfig.
func EncryptionKeySecretName(cr v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-%s", ClusterID(&cr), "encryption")
}

func IsProviderSpecForAWS(cluster v1alpha1.Cluster) bool {
	_, err := g8sClusterSpecFromCMAClusterSpec(cluster.Spec.ProviderSpec)
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
