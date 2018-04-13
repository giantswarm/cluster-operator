package chart

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseVersion, err := r.apprClient.GetReleaseVersion(chartOperatorChart, chartOperatorChannel)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ResourceState{
		ChartName:      chartOperatorChart,
		ReleaseName:    chartOperatorRelease,
		ReleaseVersion: releaseVersion,
	}

	return chartState, nil
}
