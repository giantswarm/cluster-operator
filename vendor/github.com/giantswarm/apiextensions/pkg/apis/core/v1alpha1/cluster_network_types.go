package v1alpha1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewClusterNetworkConfigCRD returns a new custom resource definition for ClusterNetworkConfig.
// This might look something like the following.
//
//     apiVersion: apiextensions.k8s.io/v1beta1
//     kind: CustomResourceDefinition
//     metadata:
//       name: clusternetworkconfigs.core.giantswarm.io
//     spec:
//       group: core.giantswarm.io
//       scope: Namespaced
//       version: v1alpha1
//       names:
//         kind: ClusterNetworkConfig
//         plural: clusternetworkconfigs
//         singular: clusternetworkconfig
//       # subresources describes the subresources for custom resource.
//       subresources:
//          # status enables the status subresource.
//         status: {}
//
func NewClusterNetworkConfigCRD() *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiextensionsv1beta1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "clusternetworkconfigs.core.giantswarm.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "core.giantswarm.io",
			Scope:   "Namespaced",
			Version: "v1alpha1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:     "ClusterNetworkConfig",
				Plural:   "clusternetworkconfigs",
				Singular: "clusternetworkconfig",
			},
			Subresources: &apiextensionsv1beta1.CustomResourceSubresources{
				Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
			},
		},
	}
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterNetworkConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterNetworkConfigSpec   `json:"spec" yaml:"spec"`
	Status            ClusterNetworkConfigStatus `json:"status" yaml:"status"`
}

type ClusterNetworkConfigSpec struct {
	MaskBits      int                                   `json:"maskBits" yaml:"maskBits"`
	VersionBundle ClusterNetworkConfigSpecVersionBundle `json:"versionBundle" yaml:"versionBundle"`
}

type ClusterNetworkConfigSpecVersionBundle struct {
	Version string `json:"version" yaml:"version"`
}

type ClusterNetworkConfigStatus struct {
	IP   string `json:"ip" yaml:"ip"`
	Mask string `json:"mask" yaml:"mask"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterNetworkConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ClusterNetworkConfig `json:"items"`
}
