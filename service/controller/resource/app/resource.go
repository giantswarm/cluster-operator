package app

import (
	"github.com/ghodss/yaml"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/service/internal/basedomain"
)

const (
	Name = "app"
)

// Config represents the configuration used to create a new chartconfig service.
type Config struct {
	BaseDomain basedomain.Interface
	G8sClient  versioned.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	Provider             string
	RawAppDefaultConfig  string
	RawAppOverrideConfig string
}

// Resource provides shared functionality for managing chartconfigs.
type Resource struct {
	baseDomain basedomain.Interface
	g8sClient  versioned.Interface
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger

	defaultConfig  defaultConfig
	overrideConfig overrideConfig
	provider       string
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
	if config.BaseDomain == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
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
		baseDomain: config.BaseDomain,
		g8sClient:  config.G8sClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,

		defaultConfig:  defaultConfig,
		overrideConfig: overrideConfig,
		provider:       config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
