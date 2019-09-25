package app

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	Name = "appv20"
)

// Config represents the configuration used to create a new chartconfig service.
type Config struct {
	G8sClient                versioned.Interface
	GetClusterConfigFunc     func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	GetClusterObjectMetaFunc func(obj interface{}) (metav1.ObjectMeta, error)
	K8sClient                kubernetes.Interface
	Logger                   micrologger.Logger

	Provider string
}

// Resource provides shared functionality for managing chartconfigs.
type Resource struct {
	g8sClient                versioned.Interface
	getClusterConfigFunc     func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	getClusterObjectMetaFunc func(obj interface{}) (metav1.ObjectMeta, error)
	k8sClient                kubernetes.Interface
	logger                   micrologger.Logger

	provider string
}

// New creates a new chartconfig service.
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
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &Resource{
		g8sClient:                config.G8sClient,
		getClusterConfigFunc:     config.GetClusterConfigFunc,
		getClusterObjectMetaFunc: config.GetClusterObjectMetaFunc,
		k8sClient:                config.K8sClient,
		logger:                   config.Logger,

		provider: config.Provider,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
