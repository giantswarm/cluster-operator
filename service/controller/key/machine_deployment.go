package key

import (
	"github.com/giantswarm/microerror"
	clusterv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

func ToMachineDeployment(v interface{}) (clusterv1alpha2.MachineDeployment, error) {
	if v == nil {
		return clusterv1alpha2.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &clusterv1alpha2.MachineDeployment{}, v)
	}

	p, ok := v.(*clusterv1alpha2.MachineDeployment)
	if !ok {
		return clusterv1alpha2.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &clusterv1alpha2.MachineDeployment{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
