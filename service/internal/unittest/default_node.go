package unittest

import (
	//"time"

	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultNodes() corev1.NodeList {
	workerNode := NewWorkerNode()
	masterNode := NewMasterNode()
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
		Items: []corev1.Node{workerNode, masterNode},
	}
	return nodes
}

func NewWorkerNode() corev1.Node {
	n := corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ip-10-0-5-28.eu-central-1.compute.internal",
			Labels: map[string]string{
				"node-role.kubernetes.io/worker":   "",
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
	}
	return n
}

func NewMasterNode() corev1.Node {
	n := corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ip-10-0-5-29.eu-central-1.compute.internal",
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

func NewAdditionalMasterNode() corev1.Node {
	n := corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ip-10-0-5-50.eu-central-1.compute.internal",
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
