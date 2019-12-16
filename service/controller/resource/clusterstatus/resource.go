package clusterstatus

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "clusterstatus"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	NewCommonClusterObject func() infrastructurev1alpha2.CommonClusterObject
	Provider               string
}

type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	newCommonClusterObject func() infrastructurev1alpha2.CommonClusterObject
	provider               string
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.NewCommonClusterObject == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.NewCommonClusterObject must not be empty", config)
	}
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		newCommonClusterObject: config.NewCommonClusterObject,
		provider:               config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
