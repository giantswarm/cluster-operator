package chartoperator

import (
	"context"

	"github.com/giantswarm/operatorkit/controller"
)

// ApplyDeleteChange is a no-op because chart-operator in the tenant cluster is
// deleted with the tenant cluster itself.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

// ApplyDeleteChange is a no-op because chart-operator in the tenant cluster is
// deleted with the tenant cluster itself.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}
