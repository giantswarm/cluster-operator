// Package label contains common Kubernetes object labels. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md
package label

const (
	ClusterIDLabel = "giantswarm.io/cluster-id"
	ManagedByLabel = "giantswarm.io/managed-by"
)
