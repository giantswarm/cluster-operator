package app

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	Name = "app"
)

// Config represents the configuration used to create a new chartconfig service.
type Config struct {
	ClusterClient *clusterclient.Client
	G8sClient     versioned.Interface
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger

	Provider string
}

// Resource provides shared functionality for managing chartconfigs.
type Resource struct {
	clusterClient *clusterclient.Client
	g8sClient     versioned.Interface
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger

	provider string
}

// New creates a new chartconfig service.
func New(config Config) (*Resource, error) {
	if config.ClusterClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		clusterClient: config.ClusterClient,
		g8sClient:     config.G8sClient,
		k8sClient:     config.K8sClient,
		logger:        config.Logger,

		provider: config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
