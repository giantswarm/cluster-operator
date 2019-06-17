package v1alpha1

// CommonClusterStatus is shared type to contain provider independent cluster status
// information.
type CommonClusterStatus struct {
	Conditions []CommonClusterStatusCondition `json:"conditions" yaml:"conditions"`
	ID         string                         `json:"id" yaml:"id"`
	Versions   []CommonClusterStatusVersion   `json:"versions" yaml:"versions"`
}

type CommonClusterStatusCondition struct {
	LastTransitionTime DeepCopyTime `json:"lastTransitionTime" yaml:"lastTransitionTime"`
	Type               string       `json:"type" yaml:"type"`
}

type CommonClusterStatusVersion struct {
	LastTransitionTime DeepCopyTime `json:"lastTransitionTime" yaml:"lastTransitionTime"`
	Version            string       `json:"version" yaml:"version"`
}
