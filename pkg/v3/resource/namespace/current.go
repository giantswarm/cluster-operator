package namespace

import (
	"context"
	"fmt"

	"github.com/giantswarm/cluster-operator/pkg/v3/guestcluster"
	"github.com/giantswarm/cluster-operator/pkg/v3/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	objectMeta, err := r.toClusterObjectMetaFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("Object Meta %#v DeletionTimestamp %v IsDeleted %t", objectMeta, objectMeta.DeletionTimestamp, key.IsDeleted(objectMeta)))

	/*
		// Guest cluster namespace is not deleted so cancel the reconcilation. The
		// namespace will be deleted when the guest cluster resources are deleted.
		if key.IsDeleted(objectMeta) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling namespace deletion: deleted with the guest cluster")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil
		}
	*/

	guestK8sClient, err := r.getGuestK8sClient(ctx, obj)
	if guestcluster.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not get a K8s client for the guest cluster")

		// We can't continue without a K8s client. We will retry during the
		// next execution.
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

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
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found the namespace in the guest cluster")
			namespace = manifest
		}
	}

	return namespace, nil
}
