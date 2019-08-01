package app

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "appv18"
)

// Config represents the configuration used to create a new app resource.
type Config struct {
	K8sClient kubernetes.Interface
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	ProjectName string
}

func New(config Config) (*StateGetter, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	s := &StateGetter{
		k8sClient: config.K8sClient,
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		projectName: config.ProjectName,
	}

	return s, nil
}

type StateGetter struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	g8sClient versioned.Interface
	logger    micrologger.Logger

	projectName string
}

func (s *StateGetter) Name() string {
	return Name
}
