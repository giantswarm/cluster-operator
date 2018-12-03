package certconfig

import (
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/v9/key"
)

const (
	// Name is the identifier of the resource.
	Name = "certconfigv9"

	// listCertConfigLimit is the suggested maximum number of CertConfigs
	// returned in one List() call to K8s API. Server may choose to not support
	// this so this is not strict maximum.
	listCertConfigLimit = 25
)

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	BaseClusterConfig        cluster.Config
	G8sClient                versioned.Interface
	K8sClient                kubernetes.Interface
	Logger                   micrologger.Logger
	ProjectName              string
	Provider                 string
	ToClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	ToClusterObjectMetaFunc  func(obj interface{}) (apismetav1.ObjectMeta, error)
}

// Resource implements the cloud config resource.
type Resource struct {
	baseClusterConfig        cluster.Config
	g8sClient                versioned.Interface
	k8sClient                kubernetes.Interface
	logger                   micrologger.Logger
	projectName              string
	provider                 string
	toClusterGuestConfigFunc func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	toClusterObjectMetaFunc  func(obj interface{}) (apismetav1.ObjectMeta, error)
}

// New creates a new configured cloud config resource.
func New(config Config) (*Resource, error) {
	if reflect.DeepEqual(config.BaseClusterConfig, cluster.Config{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseClusterConfig must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}
	if config.ToClusterGuestConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterGuestConfigFunc must not be empty", config)
	}
	if config.ToClusterObjectMetaFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterObjectMetaFunc must not be empty", config)
	}

	newService := &Resource{
		baseClusterConfig:        config.BaseClusterConfig,
		g8sClient:                config.G8sClient,
		k8sClient:                config.K8sClient,
		logger:                   config.Logger,
		projectName:              config.ProjectName,
		provider:                 config.Provider,
		toClusterGuestConfigFunc: config.ToClusterGuestConfigFunc,
		toClusterObjectMetaFunc:  config.ToClusterObjectMetaFunc,
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
