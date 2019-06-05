package cluster

import (
	"net/url"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/go-resty/resty"

	"github.com/giantswarm/clusterclient/service/cluster/creator"
	"github.com/giantswarm/clusterclient/service/cluster/deleter"
	"github.com/giantswarm/clusterclient/service/cluster/lister"
	"github.com/giantswarm/clusterclient/service/cluster/searcher"
	"github.com/giantswarm/clusterclient/service/cluster/updater"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger     micrologger.Logger
	RestClient *resty.Client

	URL *url.URL
}

type Service struct {
	Creator  *creator.Service
	Deleter  *deleter.Service
	Lister   *lister.Service
	Searcher *searcher.Service
	Updater  *updater.Service
}

// New creates a new configured service object.
func New(config Config) (*Service, error) {
	var err error

	var creatorService *creator.Service
	{
		c := creator.Config{
			Logger:     config.Logger,
			RestClient: config.RestClient,
			URL:        config.URL,
		}

		creatorService, err = creator.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deleterService *deleter.Service
	{
		deleterConfig := deleter.DefaultConfig()

		deleterConfig.Logger = config.Logger
		deleterConfig.RestClient = config.RestClient
		deleterConfig.URL = config.URL

		deleterService, err = deleter.New(deleterConfig)
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

	var searcherService *searcher.Service
	{
		searcherConfig := searcher.DefaultConfig()

		searcherConfig.Logger = config.Logger
		searcherConfig.RestClient = config.RestClient
		searcherConfig.URL = config.URL

		searcherService, err = searcher.New(searcherConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var updaterService *updater.Service
	{
		updaterConfig := updater.DefaultConfig()

		updaterConfig.Logger = config.Logger
		updaterConfig.RestClient = config.RestClient
		updaterConfig.URL = config.URL

		updaterService, err = updater.New(updaterConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		Creator:  creatorService,
		Deleter:  deleterService,
		Lister:   listerService,
		Searcher: searcherService,
		Updater:  updaterService,
	}

	return newService, nil
}
