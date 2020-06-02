package object

import "context"

// Accessor defines an interface that can be implemented for different kinds of
// objects to provide abstracted access to cluster details despite of undelying
// types.
type Accessor interface {
	GetAPIEndpoint(ctx context.Context, obj interface{}) (string, error)
}
