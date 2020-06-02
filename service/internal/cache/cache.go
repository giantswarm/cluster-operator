package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

const DefaultExpiration = 5 * time.Minute

type cache struct {
	backingStore *gocache.Cache
}

// Scoped returns cache instance which scope is limited to returned instance.
func Scoped(expiration time.Duration) Interface {
	return &cache{
		backingStore: gocache.New(expiration, expiration/2),
	}
}

func (c *cache) Get(key string) (interface{}, bool) {
	return c.backingStore.Get(key)
}

func (c *cache) Put(key string, val interface{}) {
	c.backingStore.SetDefault(key, val)
}
