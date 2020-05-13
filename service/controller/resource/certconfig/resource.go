package certconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-operator/service/controller/internal/basedomain"
	"github.com/giantswarm/cluster-operator/service/controller/internal/hamaster"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

const (
	// Name is the identifier of the resource.
	Name = "certconfig"
)

const (
	// listCertConfigLimit is the suggested maximum number of CertConfigs returned
	// in one List() call to K8s API. Server may choose to not support this so
	// this is not strict maximum.
	listCertConfigLimit = 25
)

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	BaseDomain basedomain.Interface
	G8sClient  versioned.Interface
	HAMaster   hamaster.Interface
	Logger     micrologger.Logger

	APIIP         string
	CertTTL       string
	ClusterDomain string
	Provider      string
}

// Resource implements the cloud config resource.
type Resource struct {
	baseDomain basedomain.Interface
	g8sClient  versioned.Interface
	haMaster   hamaster.Interface
	logger     micrologger.Logger

	apiIP         string
	certTTL       string
	clusterDomain string
	provider      string
}

// New creates a new configured cloud config resource.
func New(config Config) (*Resource, error) {
	if config.BaseDomain == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HAMaster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HAMaster must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.APIIP == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.APIIP must not be empty", config)
	}
	if config.CertTTL == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertTTL must not be empty", config)
	}
	if config.ClusterDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterDomain must not be empty", config)
	}
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		baseDomain: config.BaseDomain,
		g8sClient:  config.G8sClient,
		haMaster:   config.HAMaster,
		logger:     config.Logger,

		apiIP:         config.APIIP,
		certTTL:       config.CertTTL,
		clusterDomain: config.ClusterDomain,
		provider:      config.Provider,
	}

	return r, nil
}

// Name returns name of the Resource.
func (r *Resource) Name() string {
	return Name
}

func containsCertConfig(list []*v1alpha1.CertConfig, item *v1alpha1.CertConfig) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func getCertConfigByName(list []*v1alpha1.CertConfig, name string) (*v1alpha1.CertConfig, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isCertConfigModified(a, b *v1alpha1.CertConfig) bool {
	aVersion := key.CertConfigCertOperatorVersion(*a)
	bVersion := key.CertConfigCertOperatorVersion(*b)
	return aVersion != bVersion
}

func toCertConfigs(v interface{}) ([]*v1alpha1.CertConfig, error) {
	if v == nil {
		return nil, nil
	}

	certConfigs, ok := v.([]*v1alpha1.CertConfig)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", certConfigs, v)
	}

	return certConfigs, nil
}
