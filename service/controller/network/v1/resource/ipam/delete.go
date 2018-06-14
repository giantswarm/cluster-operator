package ipam

import "context"

// EnsureDeleted takes care of freeing cluster subnet when ClusterNetworkConfig
// object is deleted.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
