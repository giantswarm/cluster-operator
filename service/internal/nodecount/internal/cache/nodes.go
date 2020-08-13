package cache

import (
	"context"
	"fmt"

	"github.com/giantswarm/operatorkit/v2/pkg/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

type Nodes struct {
	cache *gocache.Cache
}

func NewNodes() *Nodes {
	r := &Nodes{
		cache: gocache.New(expiration, expiration/2),
	}

	return r
}

func (r *Nodes) Get(ctx context.Context, key string) (corev1.NodeList, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val.(corev1.NodeList), true
	}

	return corev1.NodeList{}, false
}

func (r *Nodes) Key(ctx context.Context, obj metav1.Object) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, key.ClusterID(obj))
	}

	return ""
}

func (r *Nodes) Set(ctx context.Context, key string, val corev1.NodeList) {
	r.cache.SetDefault(key, val)
}
