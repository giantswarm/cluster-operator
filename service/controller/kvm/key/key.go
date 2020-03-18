package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/service/controller/key"
)

// ChartSpecs returns charts installed only for KVM.
func ChartSpecs() []key.ChartSpec {
	// Add any provider specific charts here.
	return []key.ChartSpec{}
}

// ClusterGuestConfig extracts ClusterGuestConfig from KVMClusterConfig.
func ClusterGuestConfig(kvmClusterConfig v1alpha1.KVMClusterConfig) v1alpha1.ClusterGuestConfig {
	return kvmClusterConfig.Spec.Guest.ClusterGuestConfig
}

// ClusterID extracts clusterID from v1alpha1.KVMClusterConfig.
func ClusterID(customObject v1alpha1.KVMClusterConfig) string {
	return customObject.Spec.Guest.ClusterGuestConfig.ID
}

// ToClusterGuestConfig extracts ClusterGuestConfig from KVMClusterConfig.
func ToClusterGuestConfig(kvmClusterConfig v1alpha1.KVMClusterConfig) v1alpha1.ClusterGuestConfig {
	return kvmClusterConfig.Spec.Guest.ClusterGuestConfig
}

// ToCustomObject converts value to v1alpha1.KVMClusterConfig and returns it or
// error if type does not match.
func ToCustomObject(v interface{}) (v1alpha1.KVMClusterConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.KVMClusterConfig)
	if !ok {
		return v1alpha1.KVMClusterConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.KVMClusterConfig{}, v)
	}

	if customObjectPointer == nil {
		return v1alpha1.KVMClusterConfig{}, microerror.Maskf(emptyValueError,
			"empty value cannot be converted to CustomObject")
	}

	return *customObjectPointer, nil
}

// VersionBundleVersion extracts version bundle version from KVMClusterConfig.
func VersionBundleVersion(kvmClusterConfig v1alpha1.KVMClusterConfig) string {
	return kvmClusterConfig.Spec.VersionBundle.Version
}

func WorkerCount(kvmClusterConfig v1alpha1.KVMClusterConfig) int {
	return len(kvmClusterConfig.Spec.Guest.Workers)
}

func WorkerMaxCPUCores(kvmClusterConfig v1alpha1.KVMClusterConfig) (maxCPUCores int, known bool) {
	if WorkerCount(kvmClusterConfig) == 0 {
		return 0, false
	}

	maxCPUCores = 0
	for _, w := range kvmClusterConfig.Spec.Guest.Workers {
		workerCPUCores := w.KVMClusterConfigSpecGuestNode.CPUCores
		if workerCPUCores > maxCPUCores {
			maxCPUCores = workerCPUCores
		}
	}

	return maxCPUCores, true
}

func WorkerMaxMemorySizeGB(kvmClusterConfig v1alpha1.KVMClusterConfig) (maxMemorySizeGB float64, known bool) {
	if WorkerCount(kvmClusterConfig) == 0 {
		return 0, false
	}

	maxMemorySizeGB = 0
	for _, w := range kvmClusterConfig.Spec.Guest.Workers {
		workerMemorySizeGB := w.KVMClusterConfigSpecGuestNode.MemorySizeGB
		if workerMemorySizeGB > maxMemorySizeGB {
			maxMemorySizeGB = workerMemorySizeGB
		}
	}

	return maxMemorySizeGB, true
}
