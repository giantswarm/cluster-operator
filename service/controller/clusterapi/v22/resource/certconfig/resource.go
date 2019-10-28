package certconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-operator/pkg/v22/key"
)

const (
	// Name is the identifier of the resource.
	Name = "certconfigv22"
)

const (
	// listCertConfigLimit is the suggested maximum number of CertConfigs returned
	// in one List() call to K8s API. Server may choose to not support this so
	// this is not strict maximum.
	listCertConfigLimit = 25
)

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	APIIP    string
	CertTTL  string
	Provider string
}

// Resource implements the cloud config resource.
type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger

	apiIP    string
	certTTL  string
	provider string
}

// New creates a new configured cloud config resource.
func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
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
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	newService := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		apiIP:    config.APIIP,
		certTTL:  config.CertTTL,
		provider: config.Provider,
	}

	return newService, nil
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
	aVersion := key.CertConfigVersionBundleVersion(*a)
	bVersion := key.CertConfigVersionBundleVersion(*b)
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
