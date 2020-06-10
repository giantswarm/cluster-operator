package tenantclient

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v2/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/internal/basedomain"
	"github.com/giantswarm/cluster-operator/service/internal/tenantclient"
)

type Config struct {
	K8sClient     k8sclient.Interface
	BaseDomain    basedomain.Interface
	TenantCluster tenantcluster.Interface
	Logger        micrologger.Logger
}

type TenantClient struct {
	k8sClient     k8sclient.Interface
	baseDomain    basedomain.Interface
	tenantCluster tenantcluster.Interface
	logger        micrologger.Logger
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
		tenantCluster: c.TenantCluster,
		logger:        c.Logger,
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
		restConfig, err = c.tenantCluster.NewRestConfig(ctx, key.ClusterID(cr),
			fmt.Sprintf("api.%s.k8s.%s", key.ClusterID(cr), bd))
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
		if tenantclient.IsInvalidConfig(err) {
			return nil, microerror.Mask(err)
		} else if err != nil {
			return nil, microerror.Mask(notAvailableError)
		}
	}
	return k8sClient, nil
}
