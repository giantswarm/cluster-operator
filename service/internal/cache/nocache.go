package cache

type noCache struct{}

// NoCache is no-op implementation for Cache interface.
var NoCache = &noCache{}

func (nc *noCache) Get(key string) (interface{}, bool) {
	return nil, false
}

func (nc *noCache) Put(key string, val interface{}) {
}
