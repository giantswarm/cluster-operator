package namespace

import (
	"context"
)

// EnsureDeleted is not implemented for the namespace resource. The namespace
// will be deleted when the tenant cluster resources are deleted.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
