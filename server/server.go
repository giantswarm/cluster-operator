package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/giantswarm/microerror"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/giantswarm/cluster-operator/server/endpoint"
	"github.com/giantswarm/cluster-operator/server/middleware"
	"github.com/giantswarm/cluster-operator/service"
)

// Config represents the configuration used to construct server object.
type Config struct {
	Service           *service.Service
	MicroServerConfig microserver.Config
}

// New creates a new server object with given configuration.
func New(config Config) (microserver.Server, error) {
	var err error

	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	if config.MicroServerConfig == nil || config.MicroServerConfig.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError,
			"config.MicroServerConfig or it's Logger must not be empty")
	}

	if config.Service == nil {
		return nil, microerror.Maskf(invalidConfigError,
			"config.Service must not be empty")
	}

	var middlewareCollection *middleware.Middleware
	{
		middlewareConfig := middleware.Config{
			Logger:  config.MicroServerConfig.Logger,
			Service: config.Service,
		}

		middlewareCollection, err = middleware.New(middlewareConfig)
		if err != nil {
			return nil, microerror.Maskf(err, "middleware.New")
		}
	}

	var endpointCollection *endpoint.Endpoint
	{
		endpointConfig := endpoint.Config{
			Logger:     config.MicroServerConfig.Logger,
			Middleware: middlewareCollection,
			Service:    config.Service,
		}

		endpointCollection, err = endpoint.New(endpointConfig)
		if err != nil {
			return nil, microerror.Maskf(err, "endpoint.New")
		}
	}

	newServer := &server{
		logger:       config.MicroServerConfig.Logger,
		bootOnce:     sync.Once{},
		config:       config.MicroServerConfig,
		serviceName:  config.MicroServerConfig.ServiceName,
		shutdownOnce: sync.Once{},
	}

	// Apply internals to the micro server config.
	newServer.config.Endpoints = []microserver.Endpoint{
		endpointCollection.Healthz,
	}

	newServer.config.ErrorEncoder = newServer.newErrorEncoder()

	return newServer, nil
}

type server struct {
	logger       micrologger.Logger
	bootOnce     sync.Once
	config       microserver.Config
	serviceName  string
	shutdownOnce sync.Once
}

func (s *server) Boot() {
	s.bootOnce.Do(func() {
		// Insert here custom boot logic for server/endpoint/middleware if needed.
	})
}

func (s *server) Config() microserver.Config {
	return s.config
}

func (s *server) Shutdown() {
	s.shutdownOnce.Do(func() {
		// Insert here custom shutdown logic for server/endpoint/middleware if needed.
	})
}

func (s *server) newErrorEncoder() kithttp.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		rErr := err.(microserver.ResponseError)
		uErr := rErr.Underlying()

		rErr.SetCode(microserver.CodeInternalError)
		rErr.SetMessage(uErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
