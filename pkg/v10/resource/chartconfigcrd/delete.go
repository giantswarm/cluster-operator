package chartconfigcrd

import "context"

// EnsureDeleted is not implemented for the chartconfigcrd resource.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
