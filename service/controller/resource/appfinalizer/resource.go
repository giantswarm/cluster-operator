package appfinalizer

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "appfinalizer"
)

type Config struct {
	G8sClient            versioned.Interface
	GetClusterConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	Logger               micrologger.Logger
}

type Resource struct {
	g8sClient            versioned.Interface
	getClusterConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	logger               micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.GetClusterConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetClusterConfigFunc must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		g8sClient:            config.G8sClient,
		getClusterConfigFunc: config.GetClusterConfigFunc,
		logger:               config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
