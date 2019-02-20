package tiller

import (
	"context"
)

// EnsureDeleted is not implemented for the tiller resource. Tiller will be
// deleted when the tenant cluster resources are deleted.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
