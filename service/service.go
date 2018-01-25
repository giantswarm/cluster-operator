package service

import (
	"fmt"
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/cluster-operator/flag"
)

// Config represents the configuration used to create a new service
type Config struct {
	// Dependencies
	Logger micrologger.Logger

	// Settings
	Flag  *flag.Flag
	Viper *viper.Viper

	Description string
	GitCommit   string
	Name        string
	Source      string
}

// DefaultConfig provides a default configuration to create a new service
func DefaultConfig() Config {
	return Config{}
}

// New creates a new service with given configuration
func New(config Config) (*Service, error) {
	// Dependencies
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	config.Logger.Log("debug", fmt.Sprintf("creating cluster-operator gitCommit:%s", config.GitCommit))

	// Settings
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}

	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	newService := &Service{
		// Dependencies

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service is a type providing implementation of microkit service interface
type Service struct {
	// Dependencies

	// Internals
	bootOnce sync.Once
}

// Boot starts top level service implementation
func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		// Insert here service startup logic
	})
}
