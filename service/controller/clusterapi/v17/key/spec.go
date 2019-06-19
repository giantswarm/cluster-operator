package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// CommonClusterStatusAccessor interface provides abstracted API to manipulate
// provider specific types in provider independent way.
type CommonClusterStatusAccessor interface {
	GetCommonClusterStatus(c cmav1alpha1.Cluster) v1alpha1.CommonClusterStatus
	SetCommonClusterStatus(c cmav1alpha1.Cluster, status v1alpha1.CommonClusterStatus) cmav1alpha1.Cluster
}

type DeletionTimestampGetter interface {
	GetDeletionTimestamp() *metav1.Time
}

type LabelsGetter interface {
	GetLabels() map[string]string
}
