// Package label contains common Kubernetes object labels. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package label

const (
	// App is a standard label for guest resources.
	App = "app"

	// Cluster label is a new style label for ClusterID
	Cluster = "giantswarm.io/cluster"

	// LegacyClusterID is an old style label for ClusterID
	LegacyClusterID = "clusterID"

	// LegacyClusterKey is an old style label to specify type of a secret that
	// is used for guest cluster. This is replaced by RandomKey.
	LegacyClusterKey = "clusterKey"

	// LegacyComponent is an old style label to identify which component a
	// specific CertConfig belongs to.
	LegacyComponent = "clusterComponent"

	// RandomKeyTypeEncryption is a type of randomkey secret used for guest
	// cluster.
	RandomKeyTypeEncryption = "encryption"

	// ManagedBy label denotes which operator manages corresponding resource.
	ManagedBy = "giantswarm.io/managed-by"

	// Organization label denotes guest cluster's organization ID as displayed
	// in the front-end.
	Organization = "giantswarm.io/organization"

	// ProviderAWS label specifies format for AWS provider ID
	ProviderAWS = "aws"

	// ProviderAzure label specifies format for Azure provider ID
	ProviderAzure = "azure"

	// ProviderKVM label specifies format for KVM provider ID
	ProviderKVM = "kvm"

	// RandomKey label specifies type of a secret that is used for guest
	// cluster.
	RandomKey = "giantswarm.io/randomkey"

	// ServiceType is a standard label for guest resources.
	ServiceType = "giantswarm.io/service-type"

	// ServiceTypeManaged is a label value for managed resources.
	ServiceTypeManaged = "managed"

	// ServiceTypeSystem is a label value for system resources.
	ServiceTypeSystem = "system"
)
