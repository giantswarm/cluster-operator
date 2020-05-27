package unittest

import (
	//"time"

	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultNodes() corev1.NodeList {
	nodes := corev1.NodeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Nodes",
			APIVersion: "v1",
		},
		ListMeta: metav1.ListMeta{
			SelfLink:           "",
			ResourceVersion:    "",
			Continue:           "",
			RemainingItemCount: nil,
		},
		Items: []corev1.Node{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "ip-10-0-5-124.eu-central-1.compute.internal",
					Labels: map[string]string{
						"node.kubernetes.io/master":      "",
						"node-role.kubernetes.io/master": "",
						"giantswarm.io/control-plane":    "",
					},
					ClusterName: "",
				},
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///eu-central-1b/i-0448c486fa2eda084",
				},
				Status: corev1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   "Ready",
							Status: "True",
							LastHeartbeatTime: metav1.Time{
								Time: time.Now().Add(-5 * time.Minute),
							},
							LastTransitionTime: metav1.Time{
								Time: time.Now(),
							},
							Reason:  "KubeletReady",
							Message: "kubelet is posting ready status",
						},
					},
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "ip-10-0-5-132.eu-central-1.compute.internal",
					Labels:      map[string]string{},
					ClusterName: "",
				},
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///eu-central-1b/i-0448c486fa2eda084",
				},
				Status: corev1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   "Ready",
							Status: "True",
							LastHeartbeatTime: metav1.Time{
								Time: time.Now().Add(-5 * time.Minute),
							},
							LastTransitionTime: metav1.Time{
								Time: time.Now(),
							},
							Reason:  "KubeletReady",
							Message: "kubelet is posting ready status",
						},
					},
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "ip-10-0-5-211.eu-central-1.compute.internal",
					Labels: map[string]string{
						"giantswarm.io/machine-deployment": "abc123",
					},
					ClusterName: "",
				},
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///eu-central-1b/i-0448c486fa2eda084",
				},
				Status: corev1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   "Ready",
							Status: "True",
							LastHeartbeatTime: metav1.Time{
								Time: time.Now().Add(-5 * time.Minute),
							},
							LastTransitionTime: metav1.Time{
								Time: time.Now(),
							},
							Reason:  "KubeletReady",
							Message: "kubelet is posting ready status",
						},
					},
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "ip-10-0-5-230.eu-central-1.compute.internal",
					Labels: map[string]string{
						"giantswarm.io/machine-deployment": "abc123",
					},
					ClusterName: "",
				},
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///eu-central-1b/i-0448c486fa2eda084",
				},
				Status: corev1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   "Unknown",
							Status: "Unknown",
							LastHeartbeatTime: metav1.Time{
								Time: time.Now().Add(-5 * time.Minute),
							},
							LastTransitionTime: metav1.Time{
								Time: time.Now(),
							},
							Reason:  "KubeletNotReady",
							Message: "kubelet is posting unknown status",
						},
					},
				},
			},
		},
	}
	return nodes
}

func NewNode() corev1.Node {
	n := corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ip-10-0-5-28.eu-central-1.compute.internal",
			Labels: map[string]string{
				"node-role.kubernetes.io/master": "",
				"giantswarm.io/control-plane":    "",
			},
			ClusterName: "",
		},
		Spec: corev1.NodeSpec{
			ProviderID: "aws:///eu-central-1b/i-0448c486fa2eda084",
		},
		Status: corev1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   "Ready",
					Status: "True",
					LastHeartbeatTime: metav1.Time{
						Time: time.Now().Add(-5 * time.Minute),
					},
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
					Reason:  "KubeletReady",
					Message: "kubelet is posting ready status",
				},
			},
		},
	}
	return n
}
