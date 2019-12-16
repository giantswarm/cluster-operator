package tenantclients

import (
	"context"

	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Status.Endpoint.Base == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no endpoint base in controller context yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	var restConfig *rest.Config
	{
		restConfig, err = r.tenant.NewRestConfig(ctx, key.ClusterID(&cr), key.APIEndpoint(cr, cc.Status.Endpoint.Base))
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
			RestConfig: rest.CopyConfig(restConfig),
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		cc.Client.TenantCluster.G8s = k8sClient.G8sClient()
		cc.Client.TenantCluster.K8s = k8sClient.K8sClient()
	}

	return nil
}
