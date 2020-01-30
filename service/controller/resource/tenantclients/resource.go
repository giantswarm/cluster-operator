package tenantclients

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

const (
	Name = "tenantclients"
)

type Config struct {
	Logger        micrologger.Logger
	Tenant        tenantcluster.Interface
	ToClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha3.Cluster, error)
}

type Resource struct {
	logger        micrologger.Logger
	tenant        tenantcluster.Interface
	toClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha3.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		logger:        config.Logger,
		tenant:        config.Tenant,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
