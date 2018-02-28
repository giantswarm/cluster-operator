package clusterspec

import (
	"net/url"
	"strings"

	v1alpha1core "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	v1alpha1provider "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/flag"
)

// Factory type acts as a state holder for supplemental v1alphaprovider.Cluster
// info.
type Factory struct {
	flag *flag.Flag
}

// NewFactory constructs new Factory
func NewFactory(f *flag.Flag) (*Factory, error) {
	// TODO: Validate Config fields to not be empty
	return &Factory{f}, nil
}

// New constructs an instance of provider/v1alpha1.Cluster filled with
// configured information.
func (f *Factory) New(clusterGuestConfig v1alpha1core.ClusterGuestConfig) (*v1alpha1provider.Cluster, error) {
	cluster := &v1alpha1provider.Cluster{}

	{
		calicoDomain, err := newCalicoDomain(clusterGuestConfig)
		if err != nil {
			return &v1alpha1provider.Cluster{}, microerror.Mask(err)
		}

		cluster.Calico.CIDR = f.flag.Guest.Cluster.Calico.CIDR
		cluster.Calico.Domain = calicoDomain
		cluster.Calico.MTU = f.flag.Guest.Cluster.Calico.MTU
		cluster.Calico.Subnet = f.flag.Guest.Cluster.Calico.Subnet
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
