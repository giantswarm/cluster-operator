package tenantclients

import (
	"context"

	"github.com/giantswarm/cluster-operator/service/internal/basedomain"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v2/pkg/tenantcluster"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

const (
	Name = "tenantclients"
)

type Config struct {
	BaseDomain    basedomain.Interface
	Logger        micrologger.Logger
	Tenant        tenantcluster.Interface
	ToClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error)
}

type Resource struct {
	baseDomain    basedomain.Interface
	logger        micrologger.Logger
	tenant        tenantcluster.Interface
	toClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.BaseDomain == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
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
		baseDomain:    config.BaseDomain,
		logger:        config.Logger,
		tenant:        config.Tenant,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
