package service

import (
	"fmt"
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/flag"
	"github.com/giantswarm/cluster-operator/service/healthz"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper

	Description string
	GitCommit   string
	Name        string
	Source      string
}

// New creates a new service with given configuration.
func New(config Config) (*Service, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	config.Logger.Log("debug", fmt.Sprintf("creating cluster-operator gitCommit:%s", config.GitCommit))

	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}

	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var k8sClient kubernetes.Interface
	{
		c := k8sclient.DefaultConfig()

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.Logger = config.Logger
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sClient, err = k8sclient.New(c)
		if err != nil {
			return nil, microerror.Maskf(err, "k8sclient.New")
		}
	}

	var healthzService *healthz.Service
	{
		c := healthz.Config{
			K8sClient: k8sClient,
			Logger:    config.Logger,
		}

		healthzService, err = healthz.New(c)
		if err != nil {
			return nil, microerror.Maskf(err, "healthz.New")
		}
	}

	newService := &Service{
		Healthz: healthzService,

		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	Healthz *healthz.Service

	bootOnce sync.Once
}

// Boot starts top level service implementation.
func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		// Insert here service startup logic.
	})
}
