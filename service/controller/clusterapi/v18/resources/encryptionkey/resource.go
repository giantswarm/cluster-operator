package encryptionkey

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "encryptionkeyv18"
)

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// Resource implements the cloud config resource.
type Resource struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured secret state getter resource managing encryption
// keys.
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

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
}
