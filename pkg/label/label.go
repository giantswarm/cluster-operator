// Package label contains common Kubernetes object labels. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package label

const (
	// ClusterID label for kubernetes metadata
	ClusterID = "giantswarm.io/cluster-id"

	// LegacyClusterID is an old style label for ClusterID
	LegacyClusterID = "clusterID"

	// LegacyComponent is an old style label to identify which component a
	// specific CertConfig belongs to.
	LegacyComponent = "clusterComponent"

	// ManagedBy label denotes which operator manages corresponding resource.
	ManagedBy = "giantswarm.io/managed-by"
)
