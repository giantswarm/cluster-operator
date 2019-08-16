package chartoperator

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/pkg/v19/key"
)

// GetDesiredState returns the chart that should be installed including the
// release version it gets from the CNR channel.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseVersion, err := r.apprClient.GetReleaseVersion(ctx, chartOperatorChart, chartOperatorChannel)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterDNSIP, err := key.DNSIP(r.clusterIPRange)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	values := Values{
		ClusterDNSIP: clusterDNSIP,
		Image: Image{
			Registry: r.registryDomain,
		},
		Tiller: Tiller{
			Namespace: chartOperatorNamespace,
		},
	}

	chartState := &ResourceState{
		ChartName:      chartOperatorChart,
		ChartValues:    values,
		ReleaseName:    chartOperatorRelease,
		ReleaseVersion: releaseVersion,
		ReleaseStatus:  chartOperatorDesiredStatus,
	}

	return chartState, nil
}
