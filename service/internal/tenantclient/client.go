package tenantclient

import (
	"context"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v3/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/internal/basedomain"
)

type Config struct {
	BaseDomain    basedomain.Interface
	K8sClient     k8sclient.Interface
	Logger        micrologger.Logger
	TenantCluster tenantcluster.Interface
}

type TenantClient struct {
	baseDomain    basedomain.Interface
	k8sClient     k8sclient.Interface
	logger        micrologger.Logger
	tenantCluster tenantcluster.Interface
}

func New(c Config) (*TenantClient, error) {
	if c.BaseDomain == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", c)
	}
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}
	if c.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}
	if c.TenantCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantCluster must not be empty", c)
	}

	tenantClient := &TenantClient{
		baseDomain:    c.BaseDomain,
		k8sClient:     c.K8sClient,
		logger:        c.Logger,
		tenantCluster: c.TenantCluster,
	}

	return tenantClient, nil
}

func (c *TenantClient) K8sClient(ctx context.Context, obj interface{}) (k8sclient.Interface, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	bd, err := c.baseDomain.BaseDomain(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var restConfig *rest.Config
	{
		restConfig, err = c.tenantCluster.NewRestConfig(ctx, key.ClusterID(cr), key.APIEndpoint(cr, bd))
		if tenantcluster.IsTimeout(err) {
			return nil, microerror.Mask(notAvailableError)

		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var k8sClient k8sclient.Interface
	{
		c := k8sclient.ClientsConfig{
			Logger:     c.logger,
			RestConfig: rest.CopyConfig(restConfig),
		}

		k8sClient, err = k8sclient.NewClients(c)
		if tenant.IsAPINotAvailable(err) {
			return nil, microerror.Mask(notAvailableError)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	return k8sClient, nil
}
