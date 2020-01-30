package app

import (
	"github.com/ghodss/yaml"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/flag"
)

const (
	Name = "app"
)

// Config represents the configuration used to create a new chartconfig service.
type Config struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	Flag     *flag.Flag
	Provider string
	Viper    *viper.Viper
}

// Resource provides shared functionality for managing chartconfigs.
type Resource struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

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

	rawDefaultConfig := config.Viper.GetString(config.Flag.Service.Release.App.Config.Default)
	if rawDefaultConfig == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Flag.Service.App.Config.Default must not be empty", config)
	}

	defaultConfig := defaultConfig{}
	err := yaml.Unmarshal([]byte(rawDefaultConfig), &defaultConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	rawOverrideConfig := config.Viper.GetString(config.Flag.Service.Release.App.Config.Override)
	if rawOverrideConfig == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Flag.Service.App.Config.Override must not be empty", config)
	}

	overrideConfig := overrideConfig{}
	err = yaml.Unmarshal([]byte(rawOverrideConfig), &overrideConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		defaultConfig:  defaultConfig,
		overrideConfig: overrideConfig,
		provider:       config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
