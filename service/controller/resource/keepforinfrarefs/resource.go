package keepforinfrarefs

import (
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
)

const (
	Name = "keepforinfrarefs"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	ToObjRef func(v interface{}) (corev1.ObjectReference, error)
}

// Resource implements the operatorkit resource interface to ensure the
// reconciled runtime objects are kept until the runtime objects linked in their
// infrastructure references are deleted. For instance the Cluster CR links to
// an AWSCluster CR in its infrastructure reference. When a cluster is deleted
// the Cluster CR must not be deleted as long as the AWSCluster CR exists.
type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	toObjRef func(v interface{}) (corev1.ObjectReference, error)
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

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		toObjRef: config.ToObjRef,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
