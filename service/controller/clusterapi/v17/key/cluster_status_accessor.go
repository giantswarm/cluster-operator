package key

import (
	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type AWSClusterStatusAccessor struct{}

func (a *AWSClusterStatusAccessor) GetCommonClusterStatus(c cmav1alpha1.Cluster) g8sv1alpha1.CommonClusterStatus {
	return clusterProviderStatus(c).Cluster
}

func (a *AWSClusterStatusAccessor) SetCommonClusterStatus(c cmav1alpha1.Cluster, clusterStatus g8sv1alpha1.CommonClusterStatus) cmav1alpha1.Cluster {
	status := clusterProviderStatus(c)
	status.Cluster = clusterStatus
	return setG8sClusterStatusToCMAClusterStatus(c, status)
}
