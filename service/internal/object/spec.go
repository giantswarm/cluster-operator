package object

import "context"

// Accessor defines an interface that can be implemented for different kinds of
// objects to provide abstracted access to cluster details despite of undelying
// types.
type Accessor interface {
	GetAPIEndpoint(ctx context.Context, obj interface{}) (string, error)
}

// Cache defines an interface to simple abstracted object cache. When
// implementing this interface, one must ensure that all methods work with nil
// and empty value receiver. This is to allow client code rely on this
// interface without concerns of whether or not caching is wired (enabled).
type Cache interface {
	Get(key string) (interface{}, bool)
	Put(key string, val interface{})
}
