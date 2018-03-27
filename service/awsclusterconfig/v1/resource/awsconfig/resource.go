package awsconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "awsconfigv1"
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

// New creates a new configured cloud config resource.
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

func toAwsConfig(v interface{}) (*v1alpha1.AWSConfig, error) {
	if v == nil {
		return nil, nil
	}

	awsConfig, ok := v.(*v1alpha1.AWSConfig)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.AWSConfig{}, v)
	}

	return awsConfig, nil
}
