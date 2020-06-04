package client

import (
	"context"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster/v2/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/internal/basedomain"
)

type Config struct {
	Client        k8sclient.Interface
	BaseDomain    basedomain.Interface
	TenantCluster tenantcluster.Interface
}

type Client struct {
	k8sClient     k8sclient.Interface
	baseDomain    basedomain.Interface
	tenantCluster tenantcluster.Interface
}

func New(c Config) (*Client, error) {
	if c.BaseDomain == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", c)
	}
	if c.Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}
	if c.TenantCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantCluster must not be empty", c)
	}

	client := &Client{
		baseDomain:    c.BaseDomain,
		k8sClient:     c.Client,
		tenantCluster: c.TenantCluster,
	}

	return client, nil
}

func (c *Client) K8sClient(ctx context.Context, obj interface{}) (k8sclient.Interface, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	bd, err := c.baseDomain.BaseDomain(ctx, cr)
	if err != nil {
		return nil, err
	}

	var restConfig *rest.Config
	{
		restConfig, err = c.tenantCluster.NewRestConfig(ctx, key.ClusterID(cr), key.APIEndpoint(cr, bd))
		if tenantcluster.IsTimeout(err) {
			// TODO
			//c.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")
			//c.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var k8sClient k8sclient.Interface
	{
		c := k8sclient.ClientsConfig{
			RestConfig: rest.CopyConfig(restConfig),
		}

		k8sClient, err = k8sclient.NewClients(c)
		if tenant.IsAPINotAvailable(err) {
			//TODO
			//c.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available yet")
			//c.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	return k8sClient, nil
}
