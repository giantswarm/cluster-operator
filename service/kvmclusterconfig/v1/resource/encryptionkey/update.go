package encryptionkey

import (
	"context"

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
	return nil, nil
}
