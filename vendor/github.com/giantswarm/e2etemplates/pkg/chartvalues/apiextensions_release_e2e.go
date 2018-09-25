package chartvalues

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

type APIExtensionsReleaseE2EConfig struct {
	Namespace string

	Operator      APIExtensionsReleaseE2EConfigOperator
	VersionBundle APIExtensionsReleaseE2EConfigVersionBundle
}

type APIExtensionsReleaseE2EConfigOperator struct {
	Name    string
	Version string
}

type APIExtensionsReleaseE2EConfigVersionBundle struct {
	Version string
}

// NewAPIExtensionsAWSConfigE2E renders values required by apiextensions-aws-config-e2e-chart.
func NewAPIExtensionsReleaseE2E(config APIExtensionsReleaseE2EConfig) (string, error) {
	if config.Namespace == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Namespace must not be empty", config)
	}
	if config.Operator.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Operator.Name must not be empty", config)
	}
	if config.Operator.Version == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Operator.Version must not be empty", config)
	}
	if config.VersionBundle.Version == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.VersionBundle.Version must not be empty", config)
	}

	values, err := render.Render(apiExtensionsReleaseE2ETemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return values, nil
}
