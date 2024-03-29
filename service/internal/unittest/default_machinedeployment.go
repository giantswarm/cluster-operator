package unittest

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func DefaultMachineDeployment() apiv1beta1.MachineDeployment {
	cr := apiv1beta1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineDeployment",
			APIVersion: "cluster.x-k8s.io/v1alpha3",
		},
		Status: apiv1beta1.MachineDeploymentStatus{
			ObservedGeneration:  0,
			Selector:            "",
			Replicas:            1,
			UpdatedReplicas:     2,
			ReadyReplicas:       1,
			AvailableReplicas:   1,
			UnavailableReplicas: 0,
		},
	}
	return cr
}
