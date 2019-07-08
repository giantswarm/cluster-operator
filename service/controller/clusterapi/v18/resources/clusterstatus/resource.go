package clusterstatus

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	Name = "clusterstatusv18"
)

type Config struct {
	Accessor  Accessor
	CMAClient clientset.Interface
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	accessor  Accessor
	cmaClient clientset.Interface
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Accessor == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Accessor must not be empty", config)
	}
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		accessor:  config.Accessor,
		cmaClient: config.CMAClient,
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
