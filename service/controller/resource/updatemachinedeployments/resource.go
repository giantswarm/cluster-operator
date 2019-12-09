package updatemachinedeployments

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "updatemachinedeployments"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	Provider string
}

// Resource implements the operatorkit resource interface to keep Cluster and
// MachineDeployment CR versions in sync. CR versions are defined in the object
// meta data labels, so syncing is as simple as writing the Cluster CR version
// label values to the MachineDeployment CR version labels.
type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	provider string
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		cmaClient: config.CMAClient,
		logger:    config.Logger,

		provider: config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
