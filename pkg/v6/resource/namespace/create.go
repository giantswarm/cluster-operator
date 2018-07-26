package namespace

import (
	"context"

	"github.com/giantswarm/errors/guest"
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating namespace in the guest cluster")

		guestK8sClient, err := r.getGuestK8sClient(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = guestK8sClient.CoreV1().Namespaces().Create(namespaceToCreate)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if guest.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster is not available.")

			// We should not hammer guest API if it is not available, the guest cluster
			// might be initializing. We will retry on next reconciliation loop.
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "creating namespace in the guest cluster: created")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating namespace in the guest cluster: already created")
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
