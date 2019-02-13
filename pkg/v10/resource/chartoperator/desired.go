package chartoperator

import (
	"context"

	"github.com/giantswarm/microerror"
)

// GetDesiredState returns the chart that should be installed including the
// release version it gets from the CNR channel.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseVersion, err := r.apprClient.GetReleaseVersion(ctx, chartOperatorChart, chartOperatorChannel)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ResourceState{
		ChartName:      chartOperatorChart,
		ReleaseName:    chartOperatorRelease,
		ReleaseVersion: releaseVersion,
		ReleaseStatus:  chartOperatorDesiredStatus,
	}

	return chartState, nil
}
