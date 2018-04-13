package guestcluster

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Config represents the configuration used to create a new guest cluster service.
type Config struct {
	CertsSearcher certs.Interface
	Logger        micrologger.Logger
}

// Service provides functionality for connecting to guest clusters.
type Service struct {
	certsSearcher certs.Interface
	logger        micrologger.Logger
}

// New creates a new guest cluster service.
func New(config Config) (*Service, error) {
	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	newService := &Service{
		certsSearcher: config.CertsSearcher,
		logger:        config.Logger,
	}

	return newService, nil
}

// NewG8sClient returns a generated clientset for the specified guest cluster.
func (s *Service) NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	s.logger.LogCtx(ctx, "level", "debug", "message", "creating G8s client for the guest cluster")

	restConfig, err := s.newKubernetesRestConfig(ctx, clusterID, apiDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", "created G8s client for the guest cluster")

	return g8sClient, nil
}

// NewK8sClient returns a Kubernetes clientset for the specified guest cluster.
func (s *Service) NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	s.logger.LogCtx(ctx, "level", "debug", "message", "creating K8s client for the guest cluster")

	restConfig, err := s.newKubernetesRestConfig(ctx, clusterID, apiDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", "created K8s client for the guest cluster")

	return k8sClient, nil
}

func (s *Service) newKubernetesRestConfig(ctx context.Context, clusterID, apiDomain string) (*rest.Config, error) {
	s.logger.LogCtx(ctx, "level", "debug", "message", "looking for certificate to connect to the guest cluster")

	operatorCerts, err := s.certsSearcher.SearchClusterOperator(clusterID)
	if certs.IsTimeout(err) {
		return nil, microerror.Maskf(notFoundError, "cluster-operator cert not found for cluster")
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", "found certificate for connecting to the guest cluster")

	c := k8srestconfig.Config{
		Logger: s.logger,

		Address:   apiDomain,
		InCluster: false,
		TLS: k8srestconfig.TLSClientConfig{
			CAData:  operatorCerts.APIServer.CA,
			CrtData: operatorCerts.APIServer.Crt,
			KeyData: operatorCerts.APIServer.Key,
		},
	}
	restConfig, err := k8srestconfig.New(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return restConfig, nil
}
