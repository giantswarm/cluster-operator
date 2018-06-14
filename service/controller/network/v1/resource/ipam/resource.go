package ipam

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "ipamv1"
)

// Config represents the configuration used to create a new cluster network config resource.
type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// Resource implements the cluster network config resource.
type Resource struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured cluster network config resource.
func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	newService := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return newService, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
}
