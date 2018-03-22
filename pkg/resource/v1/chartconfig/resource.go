package chartconfig

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "chartconfigv1"
)

// Config represents the configuration used to create a new chart config resource.
type Config struct {
	BaseClusterConfig *cluster.Config
	G8sClient         versioned.Interface
	K8sClient         kubernetes.Interface
	Logger            micrologger.Logger
	ProjectName       string
}

// Resource implements the chart config resource.
type Resource struct {
	baseClusterConfig *cluster.Config
	g8sClient         versioned.Interface
	k8sClient         kubernetes.Interface
	logger            micrologger.Logger
	projectName       string
}

// New creates a new configured chart config resource.
func New(config Config) (*Resource, error) {
	if config.BaseClusterConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.BaseClusterConfig must not be empty")
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
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	newResource := &Resource{
		baseClusterConfig: config.BaseClusterConfig,
		g8sClient:         config.G8sClient,
		k8sClient:         config.K8sClient,
		logger:            config.Logger,
		projectName:       config.ProjectName,
	}

	return newResource, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
}
