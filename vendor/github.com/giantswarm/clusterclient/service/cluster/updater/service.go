package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-resty/resty"

	"github.com/giantswarm/microclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// Endpoint is the API endpoint of the service this client action interacts
	// with.
	Endpoint = "/v1/clusters/%s/"
	// Name is the service name being implemented. This can be used for e.g.
	// logging.
	Name = "cluster/updater"
)

// Config represents the configuration used to create a service.
type Config struct {
	Logger     micrologger.Logger
	RestClient *resty.Client

	URL *url.URL
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		Logger:     nil,
		RestClient: nil,

		URL: nil,
	}
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.RestClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RestClient must not be empty")
	}

	if config.URL == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.URL must not be empty")
	}

	newService := &Service{
		logger:     config.Logger,
		restClient: config.RestClient,

		url: config.URL,
	}

	return newService, nil
}

type Service struct {
	logger     micrologger.Logger
	restClient *resty.Client

	url *url.URL
}

func (s *Service) Update(ctx context.Context, request Request) (*Response, error) {
	u, err := s.url.Parse(fmt.Sprintf(Endpoint, request.Cluster.ID))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	s.logger.Log("debug", fmt.Sprintf("sending PATCH request to %s", u.String()), "service", Name)
	r, err := microclient.Do(ctx, s.restClient.R().SetBody(request.Cluster.Patch).SetResult(DefaultResponse()).Patch, u.String())
	if err != nil {
		return nil, microerror.Mask(err)
	}
	s.logger.Log("debug", fmt.Sprintf("received status code %d", r.StatusCode()), "service", Name)

	if r.StatusCode() == http.StatusBadRequest {
		responseError := responseError{}

		parseErr := json.Unmarshal(r.Body(), &responseError)
		if parseErr != nil {
			return nil, microerror.Maskf(invalidRequestError, string(r.Body()))
		}

		return nil, microerror.Maskf(invalidRequestError, responseError.Error)
	} else if r.StatusCode() != http.StatusOK {
		return nil, microerror.Mask(fmt.Errorf(string(r.Body())))
	}

	response := r.Result().(*Response)

	return response, nil
}
