package appversionlabel

import (
	"strings"

	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-operator/v3/service/internal/releaseversion"
)

const (
	Name = "appversionlabel"
)

type Config struct {
	G8sClient      versioned.Interface
	Logger         micrologger.Logger
	ReleaseVersion releaseversion.Interface
}

type Resource struct {
	g8sClient      versioned.Interface
	logger         micrologger.Logger
	releaseVersion releaseversion.Interface
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ReleaseVersion == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ReleaseVersion must not be empty", config)
	}

	r := &Resource{
		g8sClient:      config.G8sClient,
		logger:         config.Logger,
		releaseVersion: config.ReleaseVersion,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func replaceToEscape(from string) string {
	return strings.Replace(from, "/", "~1", -1)
}

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}
