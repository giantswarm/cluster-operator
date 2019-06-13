// Package annotation contains common Kubernetes metadata. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package annotation

const (
	// ChartOperator is used to filter annotations.
	ChartOperator = "chart-operator.giantswarm.io"

	// ForceHelmUpgrade is the name of the annotation that controls whether force
	// is used when upgrading the Helm release.
	ForceHelmUpgrade = "chart-operator.giantswarm.io/force-helm-upgrade"
)
