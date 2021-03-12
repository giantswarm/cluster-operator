// Package label contains common Kubernetes metadata. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package label

const (
	Cluster = "giantswarm.io/cluster"
	// ManagedBy label denotes which operator manages corresponding resource.
	ManagedBy = "giantswarm.io/managed-by"
	// Organization label denotes guest cluster's organization ID as displayed
	// in the front-end.
	Organization = "giantswarm.io/organization"
	// Release conveys the Giant Swarm release version.
	Release = "release.giantswarm.io/version"
)
