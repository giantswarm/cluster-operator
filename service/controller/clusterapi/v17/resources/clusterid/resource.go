package clusterid

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
)

const (
	Name = "clusteridv17"
)

type Config struct {
	CMAClient                   clientset.Interface
	CommonClusterStatusAccessor key.CommonClusterStatusAccessor
	G8sClient                   versioned.Interface
	Logger                      micrologger.Logger
}

type Resource struct {
	cmaClient                   clientset.Interface
	commonClusterStatusAccessor key.CommonClusterStatusAccessor
	g8sClient                   versioned.Interface
	logger                      micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.CommonClusterStatusAccessor == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CommonClusterStatusAccessor must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		cmaClient:                   config.CMAClient,
		commonClusterStatusAccessor: config.CommonClusterStatusAccessor,
		g8sClient:                   config.G8sClient,
		logger:                      config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
