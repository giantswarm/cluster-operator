package middleware

import (
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-operator/v3/service"
)

// Config represents the configuration used to construct middleware.
type Config struct {
	Logger  micrologger.Logger
	Service *service.Service
}

// Middleware is middleware collection.
type Middleware struct {
}

// New creates a new configured middleware.
func New(config Config) (*Middleware, error) {
	return &Middleware{}, nil
}
