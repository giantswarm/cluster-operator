// Package label contains common Kubernetes metadata. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package label

const (
	// ForceHelmUpgradeAnnotationName is the name of the annotation that
	// controls whether force is used when upgrading the Helm release.
	ForceHelmUpgradeAnnotationName = "chart-operator.giantswarm.io/force-helm-upgrade"
)
