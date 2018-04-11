package chart

import (
	"context"

	"github.com/giantswarm/microerror"
)

const (
	chartOperatorChart   = "chart-operator-chart"
	chartOperatorChannel = "0.1-beta"
	chartOperatorRelease = "chart-operator"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseVersion, err := r.apprClient.GetReleaseVersion(chartOperatorChart, chartOperatorChannel)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := State{
		ChartName:      chartOperatorChart,
		ReleaseName:    chartOperatorRelease,
		ReleaseVersion: releaseVersion,
	}

	return chartState, nil
}
