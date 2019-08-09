package awsclusterconfig

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "awsclusterconfigv18"
)

// Config represents the configuration used to create a new awsclusterconfig resource.
type Config struct {
	ClusterClient *clusterclient.Client
	G8sClient     versioned.Interface
	Logger        micrologger.Logger
}

// Resource implements the awsclusterconfig resource.
type Resource struct {
	clusterClient *clusterclient.Client
	g8sClient     versioned.Interface
	logger        micrologger.Logger
}

// New creates a new configured tiller resource.
func New(config Config) (*Resource, error) {
	if config.ClusterClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		clusterClient: config.ClusterClient,
		g8sClient:     config.G8sClient,
		logger:        config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
