package encryptionkey

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
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
	r.logger.LogCtx(ctx, "debug", "computing update patch for encryption key")

	currentSecret, err := toSecret(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredSecret, err := toSecret(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()

	if currentSecret == nil && desiredSecret != nil {
		patch.SetCreateChange(desiredSecret)
	}

	r.logger.LogCtx(ctx, "debug", "update patch for encryption key computed")

	return patch, nil
}
