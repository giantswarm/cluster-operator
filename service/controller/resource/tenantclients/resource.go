package tenantclients

import (
	"github.com/giantswarm/cluster-operator/service/internal/object"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v2/pkg/tenantcluster"
)

const (
	Name = "tenantclients"
)

type Config struct {
	Logger         micrologger.Logger
	Tenant         tenantcluster.Interface
	ObjectAccessor object.Accessor
}

type Resource struct {
	logger         micrologger.Logger
	tenant         tenantcluster.Interface
	objectAccessor object.Accessor
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}
	if config.ObjectAccessor == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ObjectAccessor must not be empty", config)
	}

	r := &Resource{
		logger:         config.Logger,
		tenant:         config.Tenant,
		objectAccessor: config.ObjectAccessor,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
