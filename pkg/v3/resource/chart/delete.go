package chart

import (
	"context"
	"reflect"

	"github.com/giantswarm/cluster-operator/pkg/v3/guestcluster"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/helm/pkg/helm"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	deleteState, err := toResourceState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	guestHelmClient, err := r.getGuestHelmClient(ctx, obj)
	if guestcluster.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not get a Helm client for the guest cluster")

		// A not found error here means that the cluster-operator certificate for
		// the current guest cluster was not found. We can't continue without a Helm
		// client. We will retry during the next execution, when the certificate
		// might be available.
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

		return nil
	} else if helmclient.IsTillerInstallationFailed(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "Tiller installation failed")

		// Tiller installation can fail during guest cluster setup. We will retry
		// on next reconciliation loop.
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	if deleteState.ReleaseName != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting chart-operator chart")

		guestHelmClient.DeleteRelease(deleteState.ReleaseName, helm.DeletePurge(true))
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted chart-operator chart")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting chart-operator chart")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentResourceState, err := toResourceState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredResourceState, err := toResourceState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if chart-operator chart has to be deleted")

	if !reflect.DeepEqual(currentResourceState, ResourceState{}) && reflect.DeepEqual(currentResourceState, desiredResourceState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chart-operator chart needs to be deleted")

		return &desiredResourceState, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chart-operator chart does not need to be deleted")
	}

	return nil, nil
}
