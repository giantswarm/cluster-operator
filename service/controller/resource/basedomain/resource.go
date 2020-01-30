package basedomain

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

const (
	Name = "basedomain"
)

type Config struct {
	Logger micrologger.Logger

	ToClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha3.Cluster, error)
}

type Resource struct {
	logger micrologger.Logger

	toClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha3.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
