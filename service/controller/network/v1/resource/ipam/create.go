package ipam

import "context"

// EnsureCreated takes care of cluster subnet allocation when
// ClusterNetworkConfig object is created.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	return nil
}
