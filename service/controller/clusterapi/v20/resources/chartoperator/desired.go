package chartoperator

import (
	"context"

	"github.com/giantswarm/microerror"
)

// GetDesiredState returns the chart that should be installed including the
// release version it gets from the CNR channel.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseVersion, err := r.apprClient.GetReleaseVersion(ctx, chart, channel)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ResourceState{
		ChartName: chart,
		ChartValues: Values{
			ClusterDNSIP: r.dnsIP,
			Image: Image{
				Registry: r.registryDomain,
			},
			Tiller: Tiller{
				Namespace: namespace,
			},
		},
		ReleaseName:    release,
		ReleaseVersion: releaseVersion,
		ReleaseStatus:  desiredStatus,
	}

	return chartState, nil
}
