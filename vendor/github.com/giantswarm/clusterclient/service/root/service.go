package root

import (
	"context"
	"fmt"
	"net/url"

	"gopkg.in/resty.v1"

	"github.com/giantswarm/microclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// Endpoint is the API endpoint of the service this client action interacts
	// with.
	Endpoint = "/"
)

// Config represents the configuration used to create a root service.
type Config struct {
	Logger     micrologger.Logger
	RestClient *resty.Client

	URL *url.URL
}

// DefaultConfig provides a default configuration to create a new root service
// by best effort.
func DefaultConfig() Config {
	return Config{
		Logger:     nil,
		RestClient: nil,

		URL: nil,
	}
}

// New creates a new configured root service.
func New(config Config) (*Service, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be nil")
	}
	if config.RestClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RestClient must not be nil")
	}

	if config.URL == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.URL must not be nil")
	}

	newService := &Service{
		logger:     config.Logger,
		restClient: config.RestClient,

		url: config.URL,
	}

	return newService, nil
}

// Service implements the root service action.
type Service struct {
	logger     micrologger.Logger
	restClient *resty.Client

	url *url.URL
}

func (s *Service) Get(ctx context.Context, request Request) (*Response, error) {
	u, err := s.url.Parse(Endpoint)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r, err := microclient.Do(ctx, s.restClient.R().SetResult(DefaultResponse()).Get, u.String())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if r.StatusCode() != 200 {
		return nil, microerror.Mask(fmt.Errorf(string(r.Body())))
	}

	response := r.Result().(*Response)

	return response, nil
}
