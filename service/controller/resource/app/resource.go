package app

import (
	"github.com/ghodss/yaml"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v4/service/internal/releaseversion"
)

const (
	// Name is the identifier of the resource.
	Name = "app"

	uniqueOperatorVersion = "0.0.0"
)

// Config represents the configuration used to create a new chartconfig service.
type Config struct {
	CtrlClient     ctrlClient.Client
	K8sClient      kubernetes.Interface
	Logger         micrologger.Logger
	ReleaseVersion releaseversion.Interface

	KiamWatchDogEnabled  bool
	Provider             string
	RawAppDefaultConfig  string
	RawAppOverrideConfig string
}

// Resource provides shared functionality for managing chartconfigs.
type Resource struct {
	ctrlClient     ctrlClient.Client
	k8sClient      kubernetes.Interface
	logger         micrologger.Logger
	releaseVersion releaseversion.Interface

	defaultConfig       defaultConfig
	kiamWatchDogEnabled bool
	overrideConfig      overrideConfig
	provider            string
}

type defaultConfig struct {
	Catalog         string `json:"catalog"`
	Namespace       string `json:"namespace"`
	UseUpgradeForce bool   `json:"useUpgradeForce"`
}

type overrideProperties struct {
	Chart           string `json:"chart"`
	Namespace       string `json:"namespace"`
	UseUpgradeForce *bool  `json:"useUpgradeForce,omitempty"`
}

type overrideConfig map[string]overrideProperties

// New creates a new chartconfig service.
func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
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

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}
	if config.RawAppDefaultConfig == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RawDefaultConfig must not be empty", config)
	}
	if config.RawAppOverrideConfig == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RawOverrideConfig must not be empty", config)
	}

	defaultConfig := defaultConfig{}
	err := yaml.Unmarshal([]byte(config.RawAppDefaultConfig), &defaultConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	overrideConfig := overrideConfig{}
	err = yaml.Unmarshal([]byte(config.RawAppOverrideConfig), &overrideConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r := &Resource{
		ctrlClient:     config.CtrlClient,
		k8sClient:      config.K8sClient,
		logger:         config.Logger,
		releaseVersion: config.ReleaseVersion,

		defaultConfig:       defaultConfig,
		kiamWatchDogEnabled: config.KiamWatchDogEnabled,
		overrideConfig:      overrideConfig,
		provider:            config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
