package client

import (
	"context"

	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/internal/k8sclient/internal/cache"
)

type Config struct {
	Client k8sclient.Interface
}

type Client struct {
	k8sClient k8sclient.Interface

	clusterCache *cache.Cluster
}

func New(c Config) (*Client, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	c := &Client{
		k8sClient: c.K8sClient,

		clusterCache: cache.NewCluster(),
	}

	return c, nil
}

func (c *Client) K8sClient(ctx context.Context, obj interface{}) (k8sclient.Interface, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	client, err := c.cachedCluster(ctx, cr)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return client, nil
}

func (c *Client) cachedCluster(ctx context.Context, cr metav1.Object) (k8sclient.Interface, error) {
	var err error
	var ok bool

	var client k8sclient.Interface
	{
		ck := c.clusterCache.Key(ctx, cr)

		if ck == "" {
			client, err = c.lookupCluster(ctx, cr)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		} else {
			client, ok = c.clusterCache.Get(ctx, ck)
			if !ok {
				client, err = c.lookupCluster(ctx, cr)
				if err != nil {
					return nil, microerror.Mask(err)
				}
				// TODO
				//c.clusterCache.Set(ctx, ck, cluster)
			}
		}
	}

	return client, nil
}

func (c *Client) lookupCluster(ctx context.Context, cr metav1.Object) (k8sclient.Interface, error) {
	// TODO
	return nil, nil
}
