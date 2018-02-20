package certconfig

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "certconfigv1"
)

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	K8sClient kubernetes.Interface
	Key       Key
	Logger    micrologger.Logger
}

// Resource implements the cloud config resource.
type Resource struct {
	k8sClient kubernetes.Interface
	key       Key
	logger    micrologger.Logger
}

// New creates a new configured cloud config resource.
func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Key == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Key must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		k8sClient: config.K8sClient,
		key:       config.Key,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}
