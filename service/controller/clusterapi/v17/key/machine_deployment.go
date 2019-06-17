package key

import (
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func ToMachineDeployment(v interface{}) (v1alpha1.MachineDeployment, error) {
	if v == nil {
		return v1alpha1.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.MachineDeployment{}, v)
	}

	p, ok := v.(*v1alpha1.MachineDeployment)
	if !ok {
		return v1alpha1.MachineDeployment{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.MachineDeployment{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
