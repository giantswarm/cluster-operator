package key

import (
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

func ClusterInfraRef(cluster apiv1alpha2.Cluster) types.NamespacedName {
	return types.NamespacedName{
		Name:      cluster.Spec.InfrastructureRef.Name,
		Namespace: cluster.Spec.InfrastructureRef.Namespace,
	}
}

func MachineDeploymentInfraRef(md apiv1alpha2.MachineDeployment) types.NamespacedName {
	return types.NamespacedName{
		Name:      md.Spec.Template.Spec.InfrastructureRef.Name,
		Namespace: md.Spec.Template.Spec.InfrastructureRef.Namespace,
	}
}
