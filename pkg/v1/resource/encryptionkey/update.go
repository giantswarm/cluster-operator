package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/api/core/v1"
)

// ApplyUpdateChange takes observed custom object and update portion of the
// Patch provided by NewUpdatePatch or NewDeletePatch. This is currently a NOP
// for encryptionkey Resource.
func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	return nil
}

// NewUpdatePatch computes appropriate Patch based on difference in current
// state and desired state.
func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (*v1.Secret, error) {
	// for now encryption key should not be updated when one already exists
	return nil, nil
}
