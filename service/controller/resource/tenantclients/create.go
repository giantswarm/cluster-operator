package tenantclients

import (
	"context"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster/v2/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	m, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var restConfig *rest.Config
	{
		apiEndpoint, err := r.objectAccessor.GetAPIEndpoint(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}
		restConfig, err = r.tenant.NewRestConfig(ctx, key.ClusterID(m), apiEndpoint)
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	var k8sClient k8sclient.Interface
	{
		c := k8sclient.ClientsConfig{
			Logger:     r.logger,
			RestConfig: rest.CopyConfig(restConfig),
		}

		k8sClient, err = k8sclient.NewClients(c)
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		cc.Client.TenantCluster.G8s = k8sClient.G8sClient()
		cc.Client.TenantCluster.K8s = k8sClient.K8sClient()
	}

	return nil
}
