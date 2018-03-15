package clusterspec

import (
	"net/url"
	"strings"

	v1alpha1core "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	v1alpha1provider "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
)

// Factory type acts as a state holder for supplemental v1alphaprovider.Cluster
// info.
type Factory struct {
	baseCluster *v1alpha1provider.Cluster
}

// NewFactory constructs new Factory
func NewFactory(baseCluster *v1alpha1provider.Cluster) (*Factory, error) {
	// TODO: Validate relevant Cluster fields to not be empty
	return &Factory{baseCluster}, nil
}

// New constructs an instance of provider/v1alpha1.Cluster filled with
// configured information.
func (f *Factory) New(clusterGuestConfig v1alpha1core.ClusterGuestConfig) (*v1alpha1provider.Cluster, error) {
	cluster := &v1alpha1provider.Cluster{}

	// Deep copy base information
	f.baseCluster.DeepCopyInto(cluster)

	{
		calicoDomain, err := newCalicoDomain(clusterGuestConfig)
		if err != nil {
			return &v1alpha1provider.Cluster{}, microerror.Mask(err)
		}

		// NOTE: CIDR, MTU & Subnet are copied from pre-configured baseCluster.
		cluster.Calico.Domain = calicoDomain
	}

	// TODO: Implement rest of New() - mostly along the lines of
	// https://github.com/giantswarm/kubernetesd/blob/master/service/creator/cluster.go#L39.

	return cluster, nil
}

func newCalicoDomain(clusterGuestConfig v1alpha1core.ClusterGuestConfig) (string, error) {
	u, err := url.Parse(clusterGuestConfig.API.Endpoint)
	if err != nil {
		return "", microerror.Mask(err)
	}
	splitted := strings.Split(u.Host, ".")
	splitted[0] = string(certs.CalicoCert)
	calicoDomain := strings.Join(splitted, ".")

	return calicoDomain, nil
}
