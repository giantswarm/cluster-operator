package clusterconfigmap

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "clusterconfigmap"
)

// Config represents the configuration used to create a new clusterConfigMap resource.
type Config struct {
	// Dependencies.
	GetClusterConfigFunc         func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	GetClusterObjectMetaFunc     func(obj interface{}) (metav1.ObjectMeta, error)
	GetWorkerCountFunc           func(obj interface{}) (int, error)
	GetWorkerMaxCPUCoresFunc     func(obj interface{}) (maxCPUCores int, known bool, err error)
	GetWorkerMaxMemorySizeGBFunc func(obj interface{}) (maxMemorySizeGB float64, known bool, err error)
	K8sClient                    kubernetes.Interface
	Logger                       micrologger.Logger

	// Settings.
	CalicoAddress      string
	CalicoPrefixLength string
	ClusterIPRange     string
	Provider           string
}

// Resource implements the clusterConfigMap resource.
type StateGetter struct {
	// Dependencies.
	getClusterConfigFunc         func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	getClusterObjectMetaFunc     func(obj interface{}) (metav1.ObjectMeta, error)
	getWorkerCountFunc           func(obj interface{}) (int, error)
	getWorkerMaxCPUCoresFunc     func(obj interface{}) (maxCPUCores int, known bool, err error)
	getWorkerMaxMemorySizeGBFunc func(obj interface{}) (maxMemorySizeGB float64, known bool, err error)
	k8sClient                    kubernetes.Interface
	logger                       micrologger.Logger

	// Settings.
	calicoAddress      string
	calicoPrefixLength string
	clusterIPRange     string
	provider           string
}

// New creates a new configured clusterConfigMap resource.
func New(config Config) (*StateGetter, error) {
	// Dependencies.
	if config.GetClusterConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetClusterConfigFunc must not be empty", config)
	}
	if config.GetClusterObjectMetaFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetClusterObjectMetaFunc must not be empty", config)
	}
	if config.GetWorkerCountFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetWorkerCountFunc must not be empty", config)
	}
	if config.GetWorkerMaxCPUCoresFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetWorkerMaxCPUCoresFunc must not be empty", config)
	}
	if config.GetWorkerMaxMemorySizeGBFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetWorkerMaxMemorySizeGBFunc must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Settings.
	if config.ClusterIPRange == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterIPRange must not be empty", config)
	}
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	r := &StateGetter{
		// Dependencies.
		getClusterConfigFunc:         config.GetClusterConfigFunc,
		getClusterObjectMetaFunc:     config.GetClusterObjectMetaFunc,
		getWorkerCountFunc:           config.GetWorkerCountFunc,
		getWorkerMaxCPUCoresFunc:     config.GetWorkerMaxCPUCoresFunc,
		getWorkerMaxMemorySizeGBFunc: config.GetWorkerMaxMemorySizeGBFunc,
		k8sClient:                    config.K8sClient,
		logger:                       config.Logger,

		// Settings
		calicoAddress:      config.CalicoAddress,
		calicoPrefixLength: config.CalicoPrefixLength,
		clusterIPRange:     config.ClusterIPRange,
		provider:           config.Provider,
	}

	return r, nil
}

func (r *StateGetter) Name() string {
	return Name
}
