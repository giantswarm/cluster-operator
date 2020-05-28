package object

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

const (
	expiration = 5 * time.Minute

	objectCacheKey = "objectCache"
)

type cache struct {
	backingStore *gocache.Cache
}

func NewCache() Cache {
	return &cache{
		backingStore: gocache.New(expiration, expiration/2),
	}
}

func CacheFromContext(ctx context.Context) Cache {
	c, _ := ctx.Value(objectCacheKey).(Cache)
	return c
}

func ContextWithCache(ctx context.Context, c Cache) context.Context {
	return context.WithValue(ctx, objectCacheKey, c)
}

func (c *cache) Get(key string) (interface{}, bool) {
	if c == nil {
		return nil, false
	}

	return c.backingStore.Get(key)
}

func (c *cache) Put(key string, val interface{}) {
	if c == nil {
		return
	}

	c.backingStore.SetDefault(key, val)
}
