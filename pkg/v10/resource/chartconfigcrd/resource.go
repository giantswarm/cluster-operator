package chartconfigcrd

import (
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
)

const (
	Name = "chartconfigcrdv10"
)

// Config represents the configuration used to create a new chartconfigcrd
// resource.
type Config struct {
	BaseClusterConfig        cluster.Config
	Logger                   micrologger.Logger
	Tenant                   tenantcluster.Interface
	ToClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
}

// Resource implements the chartconfigcrd resource.
type Resource struct {
	baseClusterConfig        cluster.Config
	logger                   micrologger.Logger
	tenant                   tenantcluster.Interface
	toClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
}

// New creates a new configured chartconfigcrd resource.
func New(config Config) (*Resource, error) {
	if reflect.DeepEqual(config.BaseClusterConfig, cluster.Config{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseClusterConfig must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}
	if config.ToClusterGuestConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterGuestConfigFunc must not be empty", config)
	}

	r := &Resource{
		baseClusterConfig:        config.BaseClusterConfig,
		logger:                   config.Logger,
		tenant:                   config.Tenant,
		toClusterGuestConfigFunc: config.ToClusterGuestConfigFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
