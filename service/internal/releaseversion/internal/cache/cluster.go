package cache

import (
	"context"
	"fmt"

	"github.com/giantswarm/operatorkit/v8/pkg/controller/context/cachekeycontext"
	releasev1alpha1 "github.com/giantswarm/release-operator/v4/api/v1alpha1"
	gocache "github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
)

type Release struct {
	cache *gocache.Cache
}

func NewRelease() *Release {
	r := &Release{
		cache: gocache.New(expiration, expiration/2),
	}

	return r
}

func (r *Release) Get(ctx context.Context, key string) (releasev1alpha1.Release, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val.(releasev1alpha1.Release), true
	}

	return releasev1alpha1.Release{}, false
}

func (r *Release) Key(ctx context.Context, obj metav1.Object) string {
	rl, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", rl, key.ClusterID(obj))
	}

	return ""
}

func (r *Release) Set(ctx context.Context, key string, val releasev1alpha1.Release) {
	r.cache.SetDefault(key, val)
}
