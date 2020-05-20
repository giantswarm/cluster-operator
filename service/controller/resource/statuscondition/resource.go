package statuscondition

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-operator/service/internal/releaseversion"
)

const (
	Name = "statuscondition"
)

type Config struct {
	K8sClient     k8sclient.Interface
	Logger        micrologger.Logger
	ReleasVersion releaseversion.Interface

	NewCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
	Provider                   string
}

type Resource struct {
	k8sClient      k8sclient.Interface
	logger         micrologger.Logger
	releaseVersion releaseversion.Interface

	newCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
	provider                   string
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ReleasVersion == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ReleaseVersion must not be empty", config)
	}

	if config.NewCommonClusterObjectFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.NewCommonClusterObjectFunc must not be empty", config)
	}
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		k8sClient:      config.K8sClient,
		logger:         config.Logger,
		releaseVersion: config.ReleasVersion,

		newCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
		provider:                   config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
