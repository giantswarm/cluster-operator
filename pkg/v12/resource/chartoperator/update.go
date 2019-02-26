package chartoperator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/helm/pkg/helm"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	updateState, err := toResourceState(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if !reflect.DeepEqual(updateState, ResourceState{}) {
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
			} else if helmclient.IsTillerInstallationFailed(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "Tiller installation failed")

				// Tiller installation can fail during guest cluster setup. We will
				// retry on next reconciliation loop.
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)

				return nil
			} else if guest.IsAPINotAvailable(err) {
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
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating chart-operator chart"))

			tarballPath, err := r.apprClient.PullChartTarball(ctx, updateState.ChartName, chartOperatorChannel)
			if err != nil {
				return microerror.Mask(err)
			}
			defer func() {
				err := r.fs.Remove(tarballPath)
				if err != nil {
					r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %q failed", tarballPath), "stack", fmt.Sprintf("%#v", err))
				}
			}()

			b, err := json.Marshal(updateState.ChartValues)
			if err != nil {
				return microerror.Mask(err)
			}

			err = tenantHelmClient.UpdateReleaseFromTarball(
				ctx,
				updateState.ReleaseName,
				tarballPath,
				helm.UpdateValueOverrides(b),
			)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "updated chart-operator chart")
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not updating chart-operator chart")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentResourceState, err := toResourceState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredResourceState, err := toResourceState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if chart-operator has to be updated")

	updateState := &ResourceState{}
	if shouldUpdate(currentResourceState, desiredResourceState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chart-operator has to be updated")

		updateState = &desiredResourceState
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chart-operator does not have to be updated")
	}

	return updateState, nil
}
