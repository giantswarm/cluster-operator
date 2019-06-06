package namespace

import (
	"context"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	namespaceToCreate, err := toNamespace(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if namespaceToCreate != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating namespace in the tenant cluster")

		tenantK8sClient, err := r.getTenantK8sClient(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = tenantK8sClient.CoreV1().Namespaces().Create(namespaceToCreate)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if apierrors.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster api timeout.")

			// We should not hammer tenant API if it is not available, the tenant cluster
			// might be initializing. We will retry on next reconciliation loop.
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available.")

			// We should not hammer tenant API if it is not available, the tenant cluster
			// might be initializing. We will retry on next reconciliation loop.
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "creating namespace in the tenant cluster: created")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating namespace in the tenant cluster: already created")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentNamespace, err := toNamespace(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredNamespace, err := toNamespace(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var namespaceToCreate *apiv1.Namespace
	if currentNamespace == nil {
		namespaceToCreate = desiredNamespace
	}

	return namespaceToCreate, nil
}
