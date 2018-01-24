package operator

import (
	"sync"

	gsclient "github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

// Config represents the configuration used to construct an Operator service
type Config struct {
	// Dependencies
	G8sClient    gsclient.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	// Settings

}

// DefaultConfig provides a default configuration to create a new operator
// service
func DefaultConfig() Config {
	return Config{}
}

// Service implements the Operator service interface
type Service struct {
	// Dependencies
	logger micrologger.Logger

	// Internals
	framework *framework.Framework
	bootOnce  sync.Once
}

// New creates a new configured Operator service
func New(config Config) (*Service, error) {
	// Dependencies
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.G8sClient must not be empty")
	}

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}

	if config.K8sExtClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sExtClient must not be empty")
	}

	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings
	operatorFramework, err := newFramework(config)
	if err != nil {
		return nil, microerror.Maskf(err, "newFramework")
	}

	newService := &Service{
		// Dependencies
		logger: config.Logger,

		// Internals
		framework: operatorFramework,
		bootOnce:  sync.Once{},
	}

	return newService, nil
}

// Boot starts operator service implementation
func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		go s.framework.Boot()
	})
}
