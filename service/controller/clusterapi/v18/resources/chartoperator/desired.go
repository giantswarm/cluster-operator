package chartoperator

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/pkg/v18/key"
)

// GetDesiredState returns the chart that should be installed including the
// release version it gets from the CNR channel.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseVersion, err := r.apprClient.GetReleaseVersion(ctx, chart, channel)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// TODO we use the global key package here and not the clusterapi controller
	// specific one. This is confusing and inconsistent and should be cleaned up
	// eventually. There is no reason to compute the DNS IP over and over again.
	// It should instead be injected into the resource during programm
	// initialization.
	clusterDNSIP, err := key.DNSIP(r.clusterIPRange)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ResourceState{
		ChartName: chart,
		ChartValues: Values{
			ClusterDNSIP: clusterDNSIP,
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
