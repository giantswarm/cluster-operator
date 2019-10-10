package updatemachinedeployments

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	Name = "updatemachinedeploymentsv21"
)

type Config struct {
	CMAClient clientset.Interface
	Logger    micrologger.Logger
}

// Resource implements the operatorkit resource interface to keep Cluster and
// MachineDeployment CR versions in sync. CR versions are defined in the object
// meta data labels, so syncing is as simple as writing the Cluster CR version
// label value to the MachineDeployment CR version label value.
type Resource struct {
	cmaClient clientset.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		cmaClient: config.CMAClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
