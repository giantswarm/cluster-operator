package azureconfig

import (
	"context"

	"github.com/giantswarm/operatorkit/framework"
)

// ApplyDeleteChange takes observed custom object and delete portion of the
// Patch provided by NewUpdatePatch and NewDeletePatch. It deletes AzureConfig if
// needed.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

// NewDeletePatch is called upon observed custom object deletion. It receives
// the deleted custom object, the current state as provided by GetCurrentState
// and the desired state as provided by GetDesiredState. NewDeletePatch
// analyses the current and desired state and returns the patch to be applied by
// Create, Delete, and Update functions.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	return nil, nil
}
