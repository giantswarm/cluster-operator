package unittest

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultControlPlane() infrastructurev1alpha2.G8sControlPlane {
	cr := infrastructurev1alpha2.G8sControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "G8ControlPlane",
			APIVersion: "infrastructure.giantswarm.io/v1alpha2",
		},
		Status: infrastructurev1alpha2.G8sControlPlaneStatus{
			Replicas:      1,
			ReadyReplicas: 1,
		},
	}
	return cr
}
