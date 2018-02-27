package clusterspec

import (
	"net/url"
	"strings"

	v1alpha1core "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	v1alpha1provider "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
)

// Config holds required configuration values for factory to be able to
// construct provider independent Cluster configuration based on
// ClusterGuestConfig.
type Config struct {
	Calico struct {
		CIDR   int
		MTU    int
		Subnet string
	}

	Docker struct {
		Daemon struct {
			CIDR      string
			ExtraArgs string
		}
	}

	Etcd struct {
		AltNames string
		Port     int
		Prefix   string
	}

	Kubernetes struct {
		API struct {
			AltNames       string
			ClusterIPRange string
			InsecurePort   int
			SecurePort     int
		}

		Domain string

		Hyperkube struct {
			Docker struct {
				Image string
			}
		}

		IngressController struct {
			BaseDomain string
			Docker     struct {
				Image string
			}
			InsecurePort int
			SecurePort   int
		}

		Kubelet struct {
			AltNames string
			Labels   string
			Port     int
		}

		NetworkSetup struct {
			Docker struct {
				Image string
			}
		}

		SSH struct {
			UserList []struct {
				Name      string
				PublicKey string
			}
		}
	}

	Provider struct {
		Kind string
	}
}

// Factory to construct provider independent cluster specifications
type Factory interface {
	// New constructs provider/v1alpha1.Cluster
	New(clusterGuestConfig v1alpha1core.ClusterGuestConfig) (*v1alpha1provider.Cluster, error)
}

type factory struct {
	config *Config
}

// NewFactory constructs new Factory
func NewFactory(c *Config) (Factory, error) {
	// TODO: Validate Config fields to not be empty
	return &factory{c}, nil
}

func (f *factory) New(clusterGuestConfig v1alpha1core.ClusterGuestConfig) (*v1alpha1provider.Cluster, error) {
	cluster := &v1alpha1provider.Cluster{}

	{
		calicoDomain, err := newCalicoDomain(clusterGuestConfig)
		if err != nil {
			return &v1alpha1provider.Cluster{}, microerror.Mask(err)
		}

		cluster.Calico.CIDR = f.config.Calico.CIDR
		cluster.Calico.Domain = calicoDomain
		cluster.Calico.MTU = f.config.Calico.MTU
		cluster.Calico.Subnet = f.config.Calico.Subnet
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
