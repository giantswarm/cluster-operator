// Package label contains common Kubernetes object labels. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package label

const (
	// Cluster label is a new style label for ClusterID
	Cluster = "giantswarm.io/cluster"

	// LegacyClusterID is an old style label for ClusterID
	LegacyClusterID = "clusterID"

	// LegacyClusterKey is an old style label for cluster encryption key.
	LegacyClusterKey = "clusterKey"

	// LegacyComponent is an old style label to identify which component a
	// specific CertConfig belongs to.
	LegacyComponent = "clusterComponent"

	// LegacyEncryptionKey is an old style label for EncryptionKey
	LegacyEncryptionKey = "encryption"

	// ManagedBy label denotes which operator manages corresponding resource.
	ManagedBy = "giantswarm.io/managed-by"

	// Organization label denotes guest cluster's organization ID as displayed
	// in the front-end.
	Organization = "giantswarm.io/organization"
)
