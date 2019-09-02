package tiller

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/tenantcluster"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/key"
)

var (
	values = []string{
		"spec.template.spec.priorityClassName=giantswarm-critical",
		"spec.template.spec.tolerations[0].effect=NoSchedule",
		"spec.template.spec.tolerations[0].key=node-role.kubernetes.io/master",
		"spec.template.spec.tolerations[0].operator=Exists",
	}
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("installing tiller in tenant cluster %#q", key.ClusterID(&cr)))

		if cc.Client.TenantCluster.Helm == nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster clients not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		}

		err = cc.Client.TenantCluster.Helm.EnsureTillerInstalledWithValues(ctx, values)
		if tenantcluster.IsTimeout(err) {
			// A timeout error here means that the cluster-operator certificate for the
			// current tenant cluster was not found. We can't continue without a Helm
			// client. We will retry during the next execution, when the certificate
			// might be available.
			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
			return nil

		} else if helmclient.IsTillerNotFound(err) {
			// Tiller may not be healthy and we cannot continue without a connection to
			// Tiller. We will retry on next reconciliation loop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "no healthy tiller pod found")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
			return nil

		} else if tenant.IsAPINotAvailable(err) {
			// We should not hammer tenant API if it is not available, the tenant
			// cluster might be initializing. We will retry on next reconciliation loop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant API not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("installed tiller in tenant cluster %#q", key.ClusterID(&cr)))
	}

	return nil
}
