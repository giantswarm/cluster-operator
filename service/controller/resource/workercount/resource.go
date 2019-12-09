package workercount

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	clusterv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

const (
	Name = "workercount"
)

type Config struct {
	Logger        micrologger.Logger
	ToClusterFunc func(v interface{}) (clusterv1alpha2.Cluster, error)
}

type Resource struct {
	logger        micrologger.Logger
	toClusterFunc func(v interface{}) (clusterv1alpha2.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
