package service

import (
	"fmt"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/flag"
	"github.com/giantswarm/cluster-operator/service/healthz"
	"github.com/giantswarm/cluster-operator/service/kvm"
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

	var restConfig *rest.Config
	{
		c := k8srestconfig.DefaultConfig()

		c.Logger = config.Logger

		c.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		c.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		c.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		c.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		c.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Maskf(err, "k8srestclient.New")
		}
	}

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Maskf(err, "kubernetes.NewForConfig")
	}

	k8sExtClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Maskf(err, "apiextensionsclient.NewForConfig")
	}

	var kvmFramework *framework.Framework
	{
		c := kvm.FrameworkConfig{
			G8sClient:    g8sClient,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,

			Logger: config.Logger,
		}

		kvmFramework, err = kvm.NewFramework(c)
		if err != nil {
			return nil, microerror.Maskf(err, "kvm.NewFramework")
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
		Healthz:      healthzService,
		KVMFramework: kvmFramework,

		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	Healthz      *healthz.Service
	KVMFramework *framework.Framework

	bootOnce sync.Once
}

// Boot starts top level service implementation.
func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		go s.KVMFramework.Boot()
	})
}
