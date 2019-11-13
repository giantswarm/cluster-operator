package workercount

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "workercountv22"
)

type Config struct {
	Logger              micrologger.Logger
	ToClusterConfigFunc func(v interface{}) (v1alpha1.ClusterGuestConfig, error)
}

type Resource struct {
	logger              micrologger.Logger
	toClusterConfigFunc func(v interface{}) (v1alpha1.ClusterGuestConfig, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterConfigFunc must not be empty", config)
	}

	r := &Resource{
		logger:              config.Logger,
		toClusterConfigFunc: config.ToClusterConfigFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
