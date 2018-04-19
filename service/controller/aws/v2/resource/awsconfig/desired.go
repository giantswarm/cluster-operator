package awsconfig

import (
	"context"
)

// GetDesiredState takes observed (during create, delete and update events)
// custom object as an input and returns computed desired state for it.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	return nil, nil
}
