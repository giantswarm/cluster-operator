package encryptionkey

import (
	"context"
)

// ApplyCreateChange takes observed custom object and create portion of the
// Patch provided by NewUpdatePatch or NewDeletePatch. It creates k8s secret
// for encryption key if needed.
func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	return nil
}
