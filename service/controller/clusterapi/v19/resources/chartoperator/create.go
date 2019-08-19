package chartoperator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/giantswarm/microerror"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	createState, err := toResourceState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if !reflect.DeepEqual(createState, ResourceState{}) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating chart-operator release %#q in tenant cluster %#q", release, key.ClusterID(&cr)))

		p, err := r.apprClient.PullChartTarball(ctx, createState.ChartName, channel)
		if err != nil {
			return microerror.Mask(err)
		}
		defer func() {
			err := r.fileSystem.Remove(p)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("removing %#q failed", p), "stack", microerror.Stack(err))
			}
		}()

		b, err := json.Marshal(createState.ChartValues)
		if err != nil {
			return microerror.Mask(err)
		}

		err = cc.Client.TenantCluster.Helm.InstallReleaseFromTarball(
			ctx,
			p,
			namespace,
			helm.InstallWait(true),
			helm.ReleaseName(createState.ReleaseName),
			helm.ValueOverrides(b),
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created chart-operator release %#q in tenant cluster %#q", release, key.ClusterID(&cr)))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not create chart-operator release %#q in tenant cluster %#q", release, key.ClusterID(&cr)))
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentResourceState, err := toResourceState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredResourceState, err := toResourceState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var createState *ResourceState

	if reflect.DeepEqual(currentResourceState, ResourceState{}) {
		createState = &desiredResourceState
	}

	return createState, nil
}
