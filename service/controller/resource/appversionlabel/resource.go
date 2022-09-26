package appversionlabel

import (
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v5/service/internal/releaseversion"
)

const (
	Name = "appversionlabel"
)

type Config struct {
	CtrlClient     ctrlClient.Client
	Logger         micrologger.Logger
	ReleaseVersion releaseversion.Interface
}

type Resource struct {
	ctrlClient     ctrlClient.Client
	logger         micrologger.Logger
	releaseVersion releaseversion.Interface
}

func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ReleaseVersion == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ReleaseVersion must not be empty", config)
	}

	r := &Resource{
		ctrlClient:     config.CtrlClient,
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
