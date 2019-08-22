package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/controllercontext"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	chartConfigs, err := toChartConfigs(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigs) > 0 {
		for _, chartConfig := range chartConfigs {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating chartconfig %#q in namespace %#q", chartConfig.Name, chartConfig.Namespace))

			_, err := cc.Client.TenantCluster.G8s.CoreV1alpha1().ChartConfigs(chartConfig.Namespace).Create(chartConfig)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created chartconfig %#q in namespace %#q", chartConfig.Name, chartConfig.Namespace))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not create chartconfigs")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.ChartConfig, error) {
	currentChartConfigs, err := toChartConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigs, err := toChartConfigs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var chartConfigsToCreate []*v1alpha1.ChartConfig

	for _, desiredChartConfig := range desiredChartConfigs {
		if !containsChartConfig(currentChartConfigs, desiredChartConfig) {
			chartConfigsToCreate = append(chartConfigsToCreate, desiredChartConfig)
		}
	}

	return chartConfigsToCreate, nil
}
