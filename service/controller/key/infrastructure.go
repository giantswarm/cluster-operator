package key

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

func ObjRefFromCluster(cl apiv1alpha2.Cluster) corev1.ObjectReference {
	return *cl.Spec.InfrastructureRef
}

func ObjRefFromMachineDeployment(md apiv1alpha2.MachineDeployment) corev1.ObjectReference {
	return md.Spec.Template.Spec.InfrastructureRef
}

func ObjRefToNamespacedName(ref corev1.ObjectReference) types.NamespacedName {
	return types.NamespacedName{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
}
