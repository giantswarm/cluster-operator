package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/v3/guestcluster"
	"github.com/giantswarm/cluster-operator/pkg/v3/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	objectMeta, err := r.toClusterObjectMetaFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Guest cluster namespace is not deleted so cancel the reconcilation. The
	// namespace will be deleted when the guest cluster resources are deleted.
	if key.IsDeleted(objectMeta) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling namespace deletion: deleted with the guest cluster")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	guestK8sClient, err := r.getGuestK8sClient(ctx, obj)
	if guestcluster.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not get a K8s client for the guest cluster")

		// We can't continue without a K8s client. We will retry during the
		// next execution.
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation for custom object")

		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the namespace in the guest cluster")

	// Lookup the current state of the namespace.
	var namespace *apiv1.Namespace
	{
		manifest, err := guestK8sClient.CoreV1().Namespaces().Get(namespaceName, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the namespace in the guest cluster")
			// fall through
		} else if guestcluster.IsGuestAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster is not available")

			// We can't continue without a successful K8s connection. Cluster
			// may not be up yet. We will retry during the next execution.
			reconciliationcanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation for custom object")

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
