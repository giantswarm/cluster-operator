package updateinfrarefs

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
)

const (
	Name = "updateinfrarefs"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	ToObjRef func(v interface{}) (corev1.ObjectReference, error)
	Provider string
}

// Resource implements the operatorkit resource interface to ensure the
// following version labels in our infrastructure CRs, e.g. AWSCluster
// AWSMachineDeployments.
//
//     $PROVIDER-operator.giantswarm.io/version
//     release.giantswarm.io/version
//
// The release version label is taken from the Cluster CR and propagated. The
// provider operator version label is set with the value taken from the
// controller context versions as defined for the current release. This process
// ensures to distribute the right version labels among Giant Swarm
// infrastructure CRs during Tenant Cluster upgrades.
type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
	toObjRef  func(v interface{}) (corev1.ObjectReference, error)

	provider string
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ToObjRef == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToObjRef must not be empty", config)
	}
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		toObjRef: config.ToObjRef,
		provider: config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
