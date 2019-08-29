package namespace

import (
	"context"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/tenantcluster"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/v10/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	objectMeta, err := r.toClusterObjectMetaFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Tenant cluster resources are not deleted so cancel the reconcilation. They
	// will be deleted when the tenant cluster is deleted.
	if key.IsDeleted(objectMeta) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "redirecting deletion to provider operators")
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")

		return nil, nil
	}

	tenantK8sClient, err := r.gettenantK8sClient(ctx, obj)
	if tenantcluster.IsTimeout(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not get a K8s client for the guest cluster")

		// We can't continue without a K8s client. We will retry during the
		// next execution.
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")

		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the namespace in the guest cluster")

	// Lookup the current state of the namespace.
	var namespace *apiv1.Namespace
	{
		manifest, err := tenantK8sClient.CoreV1().Namespaces().Get(namespaceName, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the namespace in the guest cluster")
			// fall through
		} else if apierrors.IsTimeout(err) || tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster is not available")

			// We can't continue without a successful K8s connection. Cluster
			// may not be up yet. We will retry during the next execution.
			reconciliationcanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")

			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found the namespace in the guest cluster")
			namespace = manifest
		}
	}

	return namespace, nil
}
