package key

import (
	"github.com/giantswarm/microerror"
	capiv1alpha3 "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
)

func ToMachinePool(v interface{}) (capiv1alpha3.MachinePool, error) {
	if v == nil {
		return capiv1alpha3.MachinePool{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &capiv1alpha3.MachinePool{}, v)
	}

	p, ok := v.(*capiv1alpha3.MachinePool)
	if !ok {
		return capiv1alpha3.MachinePool{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &capiv1alpha3.MachinePool{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
