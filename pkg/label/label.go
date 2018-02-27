// Package label contains common Kubernetes object labels. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package label

const (
	ClusterID       = "giantswarm.io/cluster-id"
	LegacyClusterID = "clusterID"
	ManagedBy       = "giantswarm.io/managed-by"
)
