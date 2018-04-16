package kvmconfig

import (
	"context"
)

// ApplyCreateChange takes observed custom object and create portion of the
// Patch provided by NewUpdatePatch or NewDeletePatch. It creates KVMConfig if
// needed.
func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	return nil
}
