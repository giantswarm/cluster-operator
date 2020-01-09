package clusterconfigmap

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "clusterconfigmap"
)

// Config represents the configuration used to create a new clusterConfigMap
// resource.
type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	CalicoCIDR     string
	ClusterIPRange string
	DNSIP          string
}

// Resource implements the clusterConfigMap resource.
type Resource struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	calicoCIDR     string
	clusterIPRange string
	dnsIP          string
}

// New creates a new configured config map state getter resource managing
// cluster config maps.
//
//     https://godoc.org/github.com/giantswarm/operatorkit/resource/k8s/secretresource#StateGetter
//
func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.CalicoCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CalicoCIDR must not be empty", config)
	}
	if config.ClusterIPRange == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterIPRange must not be empty", config)
	}
	if config.DNSIP == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DNSIP must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		calicoCIDR:     config.CalicoCIDR,
		clusterIPRange: config.ClusterIPRange,
		dnsIP:          config.DNSIP,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
