package kubeconfig

import (
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "kubeconfigv19"
)

// Config represents the configuration used to create a new kubeconfig resource.
type Config struct {
	CertSearcher certs.Interface
	K8sClient    kubernetes.Interface
	Logger       micrologger.Logger
}

// Resource implements the kubeconfig resource.
type Resource struct {
	certsSearcher certs.Interface
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger
}

// New creates a new configured secret state getter resource managing kube
// configs.
//
//     https://godoc.org/github.com/giantswarm/operatorkit/resource/k8s/secretresource#StateGetter
//
func New(config Config) (*Resource, error) {
	if config.CertSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertSearcher must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		certsSearcher: config.CertSearcher,
		k8sClient:     config.K8sClient,
		logger:        config.Logger,
	}

	return r, nil
}
