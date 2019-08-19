package chartoperator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	updateState, err := toResourceState(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if !reflect.DeepEqual(updateState, ResourceState{}) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating chart-operator release %#q in tenant cluster %#q", release, key.ClusterID(&cr)))

		p, err := r.apprClient.PullChartTarball(ctx, updateState.ChartName, channel)
		if err != nil {
			return microerror.Mask(err)
		}
		defer func() {
			err := r.fileSystem.Remove(p)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("removing %#q failed", p), "stack", microerror.Stack(err))
			}
		}()

		b, err := json.Marshal(updateState.ChartValues)
		if err != nil {
			return microerror.Mask(err)
		}

		err = cc.Client.TenantCluster.Helm.UpdateReleaseFromTarball(
			ctx,
			updateState.ReleaseName,
			p,
			helm.UpdateValueOverrides(b),
			helm.UpgradeForce(true),
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated chart-operator release %#q in tenant cluster %#q", release, key.ClusterID(&cr)))
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

	var updateState *ResourceState

	if shouldUpdate(currentResourceState, desiredResourceState) {
		updateState = &desiredResourceState
	}

	return updateState, nil
}
