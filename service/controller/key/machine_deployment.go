package key

import (
	"github.com/giantswarm/microerror"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func ToMachineDeployment(v interface{}) (apiv1alpha3.MachineDeployment, error) {
	if v == nil {
		return apiv1alpha3.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1alpha3.MachineDeployment{}, v)
	}

	p, ok := v.(*apiv1alpha3.MachineDeployment)
	if !ok {
		return apiv1alpha3.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1alpha3.MachineDeployment{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
