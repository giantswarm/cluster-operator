package tenantclients

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	Name = "tenantclientsv17"
)

type Config struct {
	CMAClient clientset.Interface
	Logger    micrologger.Logger
	Tenant    tenantcluster.Interface
}

type Resource struct {
	cmaClient clientset.Interface
	logger    micrologger.Logger
	tenant    tenantcluster.Interface
}

func New(config Config) (*Resource, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	r := &Resource{
		cmaClient: config.CMAClient,
		logger:    config.Logger,
		tenant:    config.Tenant,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
