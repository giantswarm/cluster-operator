package chart

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/cluster-operator/pkg/v3/guestcluster"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	createState, err := toResourceState(createChange)
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
	} else if err != nil {
		return microerror.Mask(err)
	}

	if !reflect.DeepEqual(createState, ResourceState{}) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating chart-operator chart")

		tarballPath, err := r.apprClient.PullChartTarball(createState.ChartName, chartOperatorChannel)
		if err != nil {
			return microerror.Mask(err)
		}
		defer func() {
			err := r.fs.Remove(tarballPath)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %q failed", tarballPath), "stack", fmt.Sprintf("%#v", err))
			}
		}()

		// We need to pass the ValueOverrides option to make the install process
		// use the default values and prevent errors on nested values.
		//
		//     {
		//      rpc error: code = Unknown desc = render error in "cnr-server-chart/templates/deployment.yaml":
		//      template: cnr-server-chart/templates/deployment.yaml:20:26:
		//      executing "cnr-server-chart/templates/deployment.yaml" at <.Values.image.reposi...>: can't evaluate field repository in type interface {}
		//     }
		//
		err = guestHelmClient.InstallFromTarball(tarballPath, chartOperatorNamespace,
			helm.ReleaseName(createState.ReleaseName),
			helm.ValueOverrides([]byte("{}")),
			helm.InstallWait(true))
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", "created chart-operator chart")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not creating chart-operator chart")
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

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if chart-operator chart has to be created")

	createState := &ResourceState{}

	// chart-operator should be created if it is not present.
	if reflect.DeepEqual(currentResourceState, ResourceState{}) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chart-operator chart needs to be created")

		createState = &desiredResourceState
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chart-operator chart does not need to be created")
	}

	return createState, nil
}
