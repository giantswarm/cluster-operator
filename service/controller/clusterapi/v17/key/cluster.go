package key

import (
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func AWSClusterConfigName(cluster v1alpha1.Cluster) string {
	return clusterProviderStatus(cluster).Cluster.ID
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

func ClusterCredentialSecretName(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.CredentialSecret.Name
}

func ClusterCredentialSecretNamespace(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Provider.CredentialSecret.Namespace
}

func ClusterDNSZone(cluster v1alpha1.Cluster) string {
	return clusterProviderSpec(cluster).Cluster.DNS.Domain
}

func ClusterID(cluster v1alpha1.Cluster) string {
	return clusterProviderStatus(cluster).Cluster.ID
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

func ClusterReleaseVersion(cluster v1alpha1.Cluster) string {
	relVer, ok := cluster.Labels[label.ReleaseKey]
	if !ok {
		panic("Cluster object is missing release version label.")
	}
	return relVer
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
