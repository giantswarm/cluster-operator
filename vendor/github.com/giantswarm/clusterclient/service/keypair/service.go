package keypair

import (
	"net/url"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/go-resty/resty"

	"github.com/giantswarm/clusterclient/service/keypair/creator"
	"github.com/giantswarm/clusterclient/service/keypair/lister"
)

// Config represents the configuration used to create a new service.
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

// New creates a new configured service object.
func New(config Config) (*Service, error) {
	var err error

	var creatorService *creator.Service
	{
		creatorConfig := creator.DefaultConfig()

		creatorConfig.Logger = config.Logger
		creatorConfig.RestClient = config.RestClient
		creatorConfig.URL = config.URL

		creatorService, err = creator.New(creatorConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var listerService *lister.Service
	{
		listerConfig := lister.DefaultConfig()

		listerConfig.Logger = config.Logger
		listerConfig.RestClient = config.RestClient
		listerConfig.URL = config.URL

		listerService, err = lister.New(listerConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		Creator: creatorService,
		Lister:  listerService,
	}

	return newService, nil
}

type Service struct {
	Creator *creator.Service
	Lister  *lister.Service
}
