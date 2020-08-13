package updateinfrarefs

import (
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/cluster-operator/v3/service/internal/releaseversion"
)

const (
	Name = "updateinfrarefs"
)

type Config struct {
	K8sClient      k8sclient.Interface
	Logger         micrologger.Logger
	ReleaseVersion releaseversion.Interface

	Provider string
	ToObjRef func(v interface{}) (corev1.ObjectReference, error)
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
	k8sClient      k8sclient.Interface
	logger         micrologger.Logger
	releaseVersion releaseversion.Interface

	provider string
	toObjRef func(v interface{}) (corev1.ObjectReference, error)
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ReleaseVersion == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ReleaseVersion must not be empty", config)
	}

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}
	if config.ToObjRef == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToObjRef must not be empty", config)
	}

	r := &Resource{
		k8sClient:      config.K8sClient,
		logger:         config.Logger,
		releaseVersion: config.ReleaseVersion,

		provider: config.Provider,
		toObjRef: config.ToObjRef,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
