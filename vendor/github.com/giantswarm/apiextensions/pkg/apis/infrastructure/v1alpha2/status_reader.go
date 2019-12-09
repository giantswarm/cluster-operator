package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type StatusReader struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            StatusReaderStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type StatusReaderStatus struct {
	Cluster CommonClusterStatus `json:"cluster" yaml:"cluster"`
}
