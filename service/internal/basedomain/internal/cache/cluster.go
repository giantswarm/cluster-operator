package cache

import (
	"context"
	"fmt"
	"os"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/operatorkit/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/key"
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

func (r *Cluster) Get(ctx context.Context, key string) (infrastructurev1alpha2.AWSCluster, bool) {
	fmt.Fprintf(os.Stderr, "GET KEY %#q\n", key)

	val, ok := r.cache.Get(key)
	if ok {
		return val.(infrastructurev1alpha2.AWSCluster), true
	}

	return infrastructurev1alpha2.AWSCluster{}, false
}

func (r *Cluster) Key(ctx context.Context, obj metav1.Object) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		fmt.Fprintf(os.Stderr, "KEY KEY %s/%s\n", ck, key.ClusterID(obj))
		return fmt.Sprintf("%s/%s", ck, key.ClusterID(obj))
	}

	return ""
}

func (r *Cluster) Set(ctx context.Context, key string, val infrastructurev1alpha2.AWSCluster) {
	fmt.Fprintf(os.Stderr, "SET KEY %#q\n", key)
	r.cache.SetDefault(key, val)
}
