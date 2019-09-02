package tcnamespace

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	{
		if cc.Client.TenantCluster.K8s == nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster clients not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil
		}

		// Tenant cluster namespace is not deleted so cancel the resource. The
		// namespace will be deleted when the tenant cluster resources are deleted.
		if key.IsDeleted(&cr) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting namespace %#q in tenant cluster %#q", namespace, key.ClusterID(&cr)))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil
		}
	}

	var ns *corev1.Namespace
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding namespace %#q in tenant cluster %#q", namespace, key.ClusterID(&cr)))

		m, err := cc.Client.TenantCluster.K8s.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found namespace %#q in tenant cluster %#q", namespace, key.ClusterID(&cr)))
			// fall through
		} else if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster api not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)

			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			ns = m
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find namespace %#q in tenant cluster %#q", namespace, key.ClusterID(&cr)))
		}
	}

	return ns, nil
}
