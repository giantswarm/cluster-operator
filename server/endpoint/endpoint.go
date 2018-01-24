package endpoint

import (
	versionendpoint "github.com/giantswarm/microendpoint/endpoint/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-operator/server/middleware"
	"github.com/giantswarm/cluster-operator/service"
)

// Config represents the configuration used to construct an endpoint
type Config struct {
	// Dependencies
	Logger     micrologger.Logger
	Middleware *middleware.Middleware
	Service    *service.Service
}

// DefaultConfig provides a default configuration to create a new endpoint
func DefaultConfig() Config {
	return Config{}
}

// New creates a new endpoint with given configuration
func New(config Config) (*Endpoint, error) {
	var err error

	var versionEndpoint *versionendpoint.Endpoint
	{
		versionConfig := versionendpoint.DefaultConfig()
		versionConfig.Logger = config.Logger
		versionConfig.Service = config.Service.Version
		versionEndpoint, err = versionendpoint.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return &Endpoint{
		Version: versionEndpoint,
	}, nil
}

// Endpoint is the endpoint collection
type Endpoint struct {
	Version *versionendpoint.Endpoint
}
