package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/v7/key"
)

// ChartSpecs returns charts installed only for AWS.
func ChartSpecs() []key.ChartSpec {
	// Add any provider specific charts here.
	return []key.ChartSpec{
		{
			AppName:           "coredns",
			ChannelName:       "0-2-stable",
			ChartName:         "kubernetes-coredns-chart",
			ConfigMapName:     "coredns-values",
			Namespace:         metav1.NamespaceSystem,
			ReleaseName:       "coredns",
			UserConfigMapName: "coredns-user-values",
		},
	}
}

// ClusterGuestConfig extracts ClusterGuestConfig from AWSClusterConfig.
func ClusterGuestConfig(awsClusterConfig v1alpha1.AWSClusterConfig) v1alpha1.ClusterGuestConfig {
	return awsClusterConfig.Spec.Guest.ClusterGuestConfig
}

// ToCustomObject converts value to v1alpha1.AWSClusterConfig and returns it or
// error if type does not match.
func ToCustomObject(v interface{}) (v1alpha1.AWSClusterConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.AWSClusterConfig)
	if !ok {
		return v1alpha1.AWSClusterConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.AWSClusterConfig{}, v)
	}

	if customObjectPointer == nil {
		return v1alpha1.AWSClusterConfig{}, microerror.Maskf(emptyValueError, "empty value cannot be converted to CustomObject")
	}

	return *customObjectPointer, nil
}

// VersionBundleVersion extracts version bundle version from AWSClusterConfig.
func VersionBundleVersion(awsClusterConfig v1alpha1.AWSClusterConfig) string {
	return awsClusterConfig.Spec.VersionBundle.Version
}

func WorkerCount(awsClusterConfig v1alpha1.AWSClusterConfig) int {
	return len(awsClusterConfig.Spec.Guest.Workers)
}
