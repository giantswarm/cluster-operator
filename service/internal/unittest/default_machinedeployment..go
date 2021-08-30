package unittest

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func DefaultMachineDeployment() apiv1alpha3.MachineDeployment {
	cr := apiv1alpha3.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineDeployment",
			APIVersion: "cluster.x-k8s.io/v1alpha3",
		},
		Status: apiv1alpha3.MachineDeploymentStatus{
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
