package lister

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
	Endpoint = "/v1/releases/"
	Name     = "release/lister"
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

type Lister struct {
	logger     micrologger.Logger
	restClient *resty.Client

	url *url.URL
}

func New(config Config) (*Lister, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.RestClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RestClient must not be empty")
	}

	if config.URL == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.URL must not be empty")
	}

	l := &Lister{
		logger:     config.Logger,
		restClient: config.RestClient,

		url: config.URL,
	}

	return l, nil
}

func (l *Lister) List(ctx context.Context) ([]Response, error) {
	var err error

	var u *url.URL
	{
		u, err = l.url.Parse(Endpoint)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var response []Response
	{
		l.logger.Log("debug", fmt.Sprintf("sending GET request to '%s'", u.String()), "service", Name)
		r, err := microclient.Do(ctx, l.restClient.R().SetResult(DefaultResponse()).Get, u.String())
		if err != nil {
			return nil, microerror.Mask(err)
		}
		l.logger.Log("debug", fmt.Sprintf("received status code '%d'", r.StatusCode()), "service", Name)

		if r.StatusCode() != http.StatusOK {
			return nil, microerror.Mask(fmt.Errorf(string(r.Body())))
		}

		response = *(r.Result().(*[]Response))
	}

	return response, nil
}
