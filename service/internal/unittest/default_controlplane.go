package unittest

import (
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultControlPlane() infrastructurev1alpha3.G8sControlPlane {
	cr := infrastructurev1alpha3.G8sControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "G8ControlPlane",
			APIVersion: "infrastructure.giantswarm.io/v1alpha3",
		},
		Status: infrastructurev1alpha3.G8sControlPlaneStatus{
			Replicas:      1,
			ReadyReplicas: 1,
		},
	}
	return cr
}
