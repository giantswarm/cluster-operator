package appmigration

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "appmigrationv21"
)

type Config struct {
	G8sClient                versioned.Interface
	GetClusterConfigFunc     func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	GetClusterObjectMetaFunc func(obj interface{}) (metav1.ObjectMeta, error)
	Logger                   micrologger.Logger
	Tenant                   tenantcluster.Interface

	Provider string
}

type Resource struct {
	g8sClient                versioned.Interface
	getClusterConfigFunc     func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	getClusterObjectMetaFunc func(obj interface{}) (metav1.ObjectMeta, error)
	logger                   micrologger.Logger
	tenant                   tenantcluster.Interface

	provider string
}

func New(config Config) (*Resource, error) {
	if config.GetClusterConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetClusterConfigFunc must not be empty", config)
	}
	if config.GetClusterObjectMetaFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetClusterObjectMetaFunc must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		g8sClient:                config.G8sClient,
		getClusterConfigFunc:     config.GetClusterConfigFunc,
		getClusterObjectMetaFunc: config.GetClusterObjectMetaFunc,
		logger:                   config.Logger,
		tenant:                   config.Tenant,

		provider: config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func getChartConfigByName(list []v1alpha1.ChartConfig, name string) (v1alpha1.ChartConfig, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return v1alpha1.ChartConfig{}, microerror.Mask(notFoundError)
}
