package azureconfig

import (
	"context"
)

// GetCurrentState takes observed custom object as an input and based on that
// information looks for current state of AzureConfig and returns it. Return
// value is of type *v1alpha1.AzureConfig.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	return nil, nil
}
