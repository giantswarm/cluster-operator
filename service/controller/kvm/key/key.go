package key

import (
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

// AppSpecs returns apps installed only for KVM.
func AppSpecs() []key.AppSpec {
	// Add any provider specific charts here.
	return []key.AppSpec{}
}
