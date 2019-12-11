package tenantclients

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

const (
	Name = "tenantclients"
)

type Config struct {
	Logger              micrologger.Logger
	Tenant              tenantcluster.Interface
	ToClusterConfigFunc func(v interface{}) (v1alpha1.ClusterGuestConfig, error)
}

type Resource struct {
	logger              micrologger.Logger
	tenant              tenantcluster.Interface
	toClusterConfigFunc func(v interface{}) (v1alpha1.ClusterGuestConfig, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}
	if config.ToClusterConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterConfigFunc must not be empty", config)
	}

	r := &Resource{
		logger:              config.Logger,
		tenant:              config.Tenant,
		toClusterConfigFunc: config.ToClusterConfigFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var tenantAPIDomain string
	{
		tenantAPIDomain, err = key.APIDomain(cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var g8sClient versioned.Interface
	var k8sClient kubernetes.Interface
	{
		tenantRestConfig, err := r.tenant.NewRestConfig(ctx, key.ClusterID(cr), tenantAPIDomain)
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		tenantClientsConfig := k8sclient.ClientsConfig{
			Logger:     r.logger,
			RestConfig: tenantRestConfig,
		}
		tenantK8sClients, err := k8sclient.NewClients(tenantClientsConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		g8sClient = tenantK8sClients.G8sClient()
		k8sClient = tenantK8sClients.K8sClient()
	}

	{
		cc.Client.TenantCluster.G8s = g8sClient
		cc.Client.TenantCluster.K8s = k8sClient
	}

	return nil
}
