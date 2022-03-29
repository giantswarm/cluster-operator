package statuscondition

import (
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-operator/v4/service/internal/recorder"
	"github.com/giantswarm/cluster-operator/v4/service/internal/releaseversion"
	"github.com/giantswarm/cluster-operator/v4/service/internal/tenantclient"
)

const (
	Name = "statuscondition"
)

type Config struct {
	Event          recorder.Interface
	K8sClient      k8sclient.Interface
	Logger         micrologger.Logger
	ReleaseVersion releaseversion.Interface
	TenantClient   tenantclient.Interface

	NewCommonClusterObjectFunc func() infrastructurev1alpha3.CommonClusterObject
	Provider                   string
}

type Resource struct {
	event          recorder.Interface
	k8sClient      k8sclient.Interface
	logger         micrologger.Logger
	releaseVersion releaseversion.Interface
	tenantClient   tenantclient.Interface

	newCommonClusterObjectFunc func() infrastructurev1alpha3.CommonClusterObject
	provider                   string
}

func New(config Config) (*Resource, error) {
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ReleaseVersion == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ReleaseVersion must not be empty", config)
	}
	if config.TenantClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TenantClient must not be empty", config)
	}

	if config.NewCommonClusterObjectFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.NewCommonClusterObjectFunc must not be empty", config)
	}
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		event:          config.Event,
		k8sClient:      config.K8sClient,
		logger:         config.Logger,
		releaseVersion: config.ReleaseVersion,
		tenantClient:   config.TenantClient,

		newCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
		provider:                   config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
