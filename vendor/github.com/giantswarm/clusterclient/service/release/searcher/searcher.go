package searcher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"gopkg.in/resty.v1"

	"github.com/giantswarm/microclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Endpoint = "/v1/releases/%s/"
	Name     = "release/searcher"
)

type Config struct {
	Logger     micrologger.Logger
	RestClient *resty.Client

	URL *url.URL
}

func DefaultConfig() Config {
	return Config{
		Logger:     nil,
		RestClient: nil,

		URL: nil,
	}
}

type Searcher struct {
	logger     micrologger.Logger
	restClient *resty.Client

	url *url.URL
}

func New(config Config) (*Searcher, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.RestClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RestClient must not be empty")
	}

	if config.URL == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.URL must not be empty")
	}

	s := &Searcher{
		logger:     config.Logger,
		restClient: config.RestClient,

		url: config.URL,
	}

	return s, nil
}

func (s *Searcher) Search(ctx context.Context, request Request) (Response, error) {
	var err error

	var u *url.URL
	{
		u, err = s.url.Parse(fmt.Sprintf(Endpoint, request.ReleaseVersion))
		if err != nil {
			return Response{}, microerror.Mask(err)
		}
	}

	var response Response
	{
		s.logger.Log("debug", fmt.Sprintf("sending GET request to '%s'", u.String()), "service", Name)
		r, err := microclient.Do(ctx, s.restClient.R().SetResult(DefaultResponse()).Get, u.String())
		if err != nil {
			return Response{}, microerror.Mask(err)
		}
		s.logger.Log("debug", fmt.Sprintf("received status code '%d'", r.StatusCode()), "service", Name)

		if r.StatusCode() == http.StatusNotFound {
			return Response{}, microerror.Mask(notFoundError)
		} else if r.StatusCode() != http.StatusOK {
			return Response{}, microerror.Mask(fmt.Errorf(string(r.Body())))
		}

		response = *(r.Result().(*Response))
	}

	return response, nil
}
