package cache

import (
	"context"
	"fmt"

	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
)

type Cluster struct {
	cache *gocache.Cache
}

func NewCluster() *Cluster {
	r := &Cluster{
		cache: gocache.New(expiration, expiration/2),
	}

	return r
}

func (r *Cluster) Get(ctx context.Context, key string) (interface{}, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val, true
	}

	return nil, false
}

func (r *Cluster) Key(ctx context.Context, obj metav1.Object) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, key.ClusterID(obj))
	}

	return ""
}

func (r *Cluster) Set(ctx context.Context, key string, val interface{}) {
	r.cache.SetDefault(key, val)
}
