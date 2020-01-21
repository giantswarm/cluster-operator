package releaseversions

import (
	"context"
)

// EnsureDeleted is not putting the operator versions into the controller
// context because we do not want to fetch the version information on delete
// events. This is to reduce eventual friction. Cluster deletion should not be
// affected only because some releases are missing or broken when fetching them
// from cluster-service. Other resources must not rely on operator version
// information on delete events.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
