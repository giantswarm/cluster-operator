package key

import (
	"github.com/giantswarm/microerror"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func ToMachineDeployment(v interface{}) (apiv1beta1.MachineDeployment, error) {
	if v == nil {
		return apiv1beta1.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1beta1.MachineDeployment{}, v)
	}

	p, ok := v.(*apiv1beta1.MachineDeployment)
	if !ok {
		return apiv1beta1.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1beta1.MachineDeployment{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
