package kubeconfig

import (
	"github.com/giantswarm/certs/v2/pkg/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster/v2/pkg/tenantcluster"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/service/internal/basedomain"
)

const (
	// Name is the identifier of the resource.
	Name = "kubeconfig"
)

// Config represents the configuration used to create a new kubeconfig resource.
type Config struct {
	BaseDomain    basedomain.Interface
	CertsSearcher certs.Interface
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
	Tenant        tenantcluster.Interface
}

// Resource implements the kubeconfig resource.
type Resource struct {
	baseDomain    basedomain.Interface
	certsSearcher certs.Interface
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger
	tenant        tenantcluster.Interface
}

// New creates a new configured secret state getter resource managing kube
// configs.
//
//     https://godoc.org/github.com/giantswarm/operatorkit/resource/k8s/secretresource#StateGetter
//
func New(config Config) (*Resource, error) {
	if config.BaseDomain == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	r := &Resource{
		baseDomain:    config.BaseDomain,
		certsSearcher: config.CertsSearcher,
		k8sClient:     config.K8sClient,
		logger:        config.Logger,
		tenant:        config.Tenant,
	}

	return r, nil
}
