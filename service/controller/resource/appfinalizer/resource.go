package appfinalizer

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Name = "appfinalizer"
)

type Config struct {
	CtrlClient ctrlClient.Client
	Logger     micrologger.Logger
}

type Resource struct {
	ctrlClient ctrlClient.Client
	logger     micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
