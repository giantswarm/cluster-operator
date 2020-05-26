package unittest

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultMasterNode() corev1.Node {
	n := corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "",
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp: &metav1.Time{
				Time: time.Time{},
			},
			DeletionGracePeriodSeconds: nil,
			Labels: map[string]string{
				"": "",
			},
			Annotations: map[string]string{
				"": "",
			},
			OwnerReferences: nil,
			Finalizers:      nil,
			ClusterName:     "",
			ManagedFields:   nil,
		},
		Spec: corev1.NodeSpec{
			PodCIDR:       "",
			PodCIDRs:      nil,
			ProviderID:    "",
			Unschedulable: false,
			Taints:        nil,
			ConfigSource: &corev1.NodeConfigSource{
				ConfigMap: &corev1.ConfigMapNodeConfigSource{
					Namespace:        "",
					Name:             "",
					UID:              "",
					ResourceVersion:  "",
					KubeletConfigKey: "",
				},
			},
			DoNotUseExternalID: "",
		},
		Status: corev1.NodeStatus{
			Capacity: map[corev1.ResourceName]resource.Quantity{
				"": {
					Format: "",
				},
			},
			Allocatable: map[corev1.ResourceName]resource.Quantity{
				"": {
					Format: "",
				},
			},
			Phase:      "",
			Conditions: nil,
			Addresses:  nil,
			DaemonEndpoints: corev1.NodeDaemonEndpoints{
				KubeletEndpoint: corev1.DaemonEndpoint{
					Port: 0,
				},
			},
			NodeInfo: corev1.NodeSystemInfo{
				MachineID:               "",
				SystemUUID:              "",
				BootID:                  "",
				KernelVersion:           "",
				OSImage:                 "",
				ContainerRuntimeVersion: "",
				KubeletVersion:          "",
				KubeProxyVersion:        "",
				OperatingSystem:         "",
				Architecture:            "",
			},
			Images:          nil,
			VolumesInUse:    nil,
			VolumesAttached: nil,
			Config: &corev1.NodeConfigStatus{
				Assigned: &corev1.NodeConfigSource{
					ConfigMap: &corev1.ConfigMapNodeConfigSource{
						Namespace:        "",
						Name:             "",
						UID:              "",
						ResourceVersion:  "",
						KubeletConfigKey: "",
					},
				},
				Active: &corev1.NodeConfigSource{
					ConfigMap: &corev1.ConfigMapNodeConfigSource{
						Namespace:        "",
						Name:             "",
						UID:              "",
						ResourceVersion:  "",
						KubeletConfigKey: "",
					},
				},
				LastKnownGood: &corev1.NodeConfigSource{
					ConfigMap: &corev1.ConfigMapNodeConfigSource{
						Namespace:        "",
						Name:             "",
						UID:              "",
						ResourceVersion:  "",
						KubeletConfigKey: "",
					},
				},
				Error: "",
			},
		},
	}
	return n
}

func DefaultWorkerNode() corev1.Node {
	n := corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "",
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp: &metav1.Time{
				Time: time.Time{},
			},
			DeletionGracePeriodSeconds: nil,
			Labels: map[string]string{
				"": "",
			},
			Annotations: map[string]string{
				"": "",
			},
			OwnerReferences: nil,
			Finalizers:      nil,
			ClusterName:     "",
			ManagedFields:   nil,
		},
		Spec: corev1.NodeSpec{
			PodCIDR:       "",
			PodCIDRs:      nil,
			ProviderID:    "",
			Unschedulable: false,
			Taints:        nil,
			ConfigSource: &corev1.NodeConfigSource{
				ConfigMap: &corev1.ConfigMapNodeConfigSource{
					Namespace:        "",
					Name:             "",
					UID:              "",
					ResourceVersion:  "",
					KubeletConfigKey: "",
				},
			},
			DoNotUseExternalID: "",
		},
		Status: corev1.NodeStatus{
			Capacity: map[corev1.ResourceName]resource.Quantity{
				"": {
					Format: "",
				},
			},
			Allocatable: map[corev1.ResourceName]resource.Quantity{
				"": {
					Format: "",
				},
			},
			Phase:      "",
			Conditions: nil,
			Addresses:  nil,
			DaemonEndpoints: corev1.NodeDaemonEndpoints{
				KubeletEndpoint: corev1.DaemonEndpoint{
					Port: 0,
				},
			},
			NodeInfo: corev1.NodeSystemInfo{
				MachineID:               "",
				SystemUUID:              "",
				BootID:                  "",
				KernelVersion:           "",
				OSImage:                 "",
				ContainerRuntimeVersion: "",
				KubeletVersion:          "",
				KubeProxyVersion:        "",
				OperatingSystem:         "",
				Architecture:            "",
			},
			Images:          nil,
			VolumesInUse:    nil,
			VolumesAttached: nil,
			Config: &corev1.NodeConfigStatus{
				Assigned: &corev1.NodeConfigSource{
					ConfigMap: &corev1.ConfigMapNodeConfigSource{
						Namespace:        "",
						Name:             "",
						UID:              "",
						ResourceVersion:  "",
						KubeletConfigKey: "",
					},
				},
				Active: &corev1.NodeConfigSource{
					ConfigMap: &corev1.ConfigMapNodeConfigSource{
						Namespace:        "",
						Name:             "",
						UID:              "",
						ResourceVersion:  "",
						KubeletConfigKey: "",
					},
				},
				LastKnownGood: &corev1.NodeConfigSource{
					ConfigMap: &corev1.ConfigMapNodeConfigSource{
						Namespace:        "",
						Name:             "",
						UID:              "",
						ResourceVersion:  "",
						KubeletConfigKey: "",
					},
				},
				Error: "",
			},
		},
	}
	return n
}
