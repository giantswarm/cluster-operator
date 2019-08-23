package creator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"gopkg.in/resty.v1"

	"github.com/giantswarm/microclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/clusterclient/service/cluster/searcher"
)

const (
	// Endpoint is the API endpoint of the service this client action interacts
	// with.
	Endpoint = "/v1/clusters/"
	// Name is the service name being implemented. This can be used for e.g.
	// logging.
	Name = "cluster/creator"
)

// Config represents the configuration used to create a creator service.
type Config struct {
	Logger     micrologger.Logger
	RestClient *resty.Client

	URL *url.URL
}

type Service struct {
	logger     micrologger.Logger
	restClient *resty.Client

	url *url.URL
}

// New creates a new configured creator service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.RestClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RestClient must not be empty")
	}

	// Settings.
	if config.URL == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.URL must not be empty")
	}

	s := &Service{
		logger:     config.Logger,
		restClient: config.RestClient,

		url: config.URL,
	}

	return s, nil
}

func (s *Service) Create(ctx context.Context, request Request) (*Response, error) {
	// At first we are going to create a new cluster resource. The result in case
	// the requested resource was created successfully will be a response
	// containing information about the location of the created resource.
	var resourceLocation string
	{
		req := s.restClient.R()
		req.SetBody(request)

		u, err := s.url.Parse(Endpoint)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		s.logger.Log("debug", fmt.Sprintf("sending POST request to %s", u.String()), "service", Name)

		res, err := microclient.Do(ctx, req.Post, u.String())
		if err != nil {
			return nil, microerror.Mask(err)
		}
		s.logger.Log("debug", fmt.Sprintf("received status code %d", res.StatusCode()), "service", Name)

		if res.StatusCode() == http.StatusBadRequest {
			responseError := responseError{}

			parseErr := json.Unmarshal(res.Body(), &responseError)
			if parseErr != nil {
				return nil, microerror.Maskf(invalidRequestError, string(res.Body()))
			}

			return nil, microerror.Maskf(invalidRequestError, responseError.Error)
		} else if res.StatusCode() != http.StatusCreated {
			return nil, microerror.Mask(fmt.Errorf(string(res.Body())))
		}

		resourceLocation = res.Header().Get("Location")
	}

	// We know the location of the created resource from the response location
	// header. Now we request it to return relevant information about the created
	// resource in our response.
	var response *Response
	{
		u, err := s.url.Parse(resourceLocation)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		s.logger.Log("debug", fmt.Sprintf("sending GET request to %s", u.String()), "service", Name)
		r, err := microclient.Do(ctx, s.restClient.R().SetResult(searcher.DefaultResponse()).Get, u.String())
		if err != nil {
			return nil, microerror.Mask(err)
		}
		s.logger.Log("debug", fmt.Sprintf("received status code %d", r.StatusCode()), "service", Name)

		if r.StatusCode() == http.StatusNotFound {
			return nil, microerror.Mask(notFoundError)
		} else if r.StatusCode() != http.StatusOK {
			return nil, microerror.Mask(fmt.Errorf(string(r.Body())))
		}

		clientResponse := r.Result().(*searcher.Response)
		response = DefaultResponse()
		response.Cluster.ID = clientResponse.ID
	}

	return response, nil
}
