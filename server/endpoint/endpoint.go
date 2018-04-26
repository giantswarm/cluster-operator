package endpoint

import (
	"github.com/giantswarm/microendpoint/endpoint/healthz"
	"github.com/giantswarm/microendpoint/endpoint/version"
	healthzservice "github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-operator/server/middleware"
	"github.com/giantswarm/cluster-operator/service"
)

// Config represents the configuration used to construct an endpoint.
type Config struct {
	Logger     micrologger.Logger
	Middleware *middleware.Middleware
	Service    *service.Service
}

// Endpoint is the endpoint collection.
type Endpoint struct {
	Healthz *healthz.Endpoint
	Version *version.Endpoint
}

// New creates a new endpoint with given configuration.
func New(config Config) (*Endpoint, error) {
	var err error

	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	if config.Service == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Service or it's Healthz descendents must not be empty")
	}

	var healthzEndpoint *healthz.Endpoint
	{
		c := healthz.DefaultConfig()
		c.Logger = config.Logger
		c.Services = []healthzservice.Service{
			config.Service.Healthz.K8s,
		}

		healthzEndpoint, err = healthz.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionEndpoint *version.Endpoint
	{
		versionConfig := version.DefaultConfig()
		versionConfig.Logger = config.Logger
		versionConfig.Service = config.Service.Version
		versionEndpoint, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	endpoint := &Endpoint{
		Healthz: healthzEndpoint,
		Version: versionEndpoint,
	}

	return endpoint, nil
}
