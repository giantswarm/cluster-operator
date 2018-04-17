package chart

import (
	"context"

	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	guestHelmClient, err := r.getGuestHelmClient(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	deleteState, err := toResourceState(deleteChange)
	if err != nil {
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

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	return nil, nil
}
