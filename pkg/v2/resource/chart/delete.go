package chart

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/helm/pkg/helm"
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
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
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

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if the %s release has to be deleted", desiredResourceState.ReleaseName))

	if !reflect.DeepEqual(currentResourceState, ResourceState{}) && reflect.DeepEqual(currentResourceState, desiredResourceState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %s release needs to be deleted", desiredResourceState.ReleaseName))

		return &desiredResourceState, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %s release does not need to be deleted", desiredResourceState.ReleaseName))
	}

	return nil, nil
}
