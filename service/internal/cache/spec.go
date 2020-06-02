package cache

// Interface defines a simple abstracted object cache. When implementing this
// interface, one must ensure that all methods work with nil and empty value
// receiver. This is to allow client code rely on this interface without
// concerns of whether or not caching is wired (enabled).
type Interface interface {
	Get(key string) (interface{}, bool)
	Put(key string, val interface{})
}
