package endpoint

import (
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

// New creates a new endpoint with given configuration.
func New(config Config) (*Endpoint, error) {
	return &Endpoint{}, nil
}

// Endpoint is the endpoint collection.
type Endpoint struct {
}
