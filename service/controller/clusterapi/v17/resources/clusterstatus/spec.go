package clusterstatus

import (
	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// Accessor interface provides abstracted API to manipulate provider specific
// types in provider independent way.
type Accessor interface {
	GetCommonClusterStatus(c cmav1alpha1.Cluster) v1alpha1.CommonClusterStatus
	SetCommonClusterStatus(c cmav1alpha1.Cluster, status v1alpha1.CommonClusterStatus) cmav1alpha1.Cluster
}
