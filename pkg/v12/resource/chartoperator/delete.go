package chartoperator

import (
	"context"
	"reflect"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/helm/pkg/helm"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	deleteState, err := toResourceState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if deleteState.ReleaseName != "" {
		var tenantHelmClient helmclient.Interface
		{
			tenantHelmClient, err = r.getTenantHelmClient(ctx, obj)
			if tenantcluster.IsTimeout(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")

				// A timeout error here means that the cluster-operator certificate
				// for the current guest cluster was not found. We can't continue
				// without a Helm client. We will retry during the next execution, when
				// the certificate might be available.
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)

				return nil
			} else if helmclient.IsTillerNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "Tiller installation failed")

				// Tiller installation can fail during guest cluster setup. We will
				// retry on next reconciliation loop.
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)

				return nil
			} else if tenant.IsAPINotAvailable(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "guest API not available")

				// We should not hammer guest API if it is not available, the guest
				// cluster might be initializing. We will retry on next reconciliation
				// loop.
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)

				return nil
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "deleting chart-operator chart")

			tenantHelmClient.DeleteRelease(ctx, deleteState.ReleaseName, helm.DeletePurge(true))
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "deleted chart-operator chart")
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not delete chart-operator chart")
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
