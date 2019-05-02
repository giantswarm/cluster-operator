package kubeconfig

import (
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "kubeconfigv16"
)

// Config represents the configuration used to create a new kubeconfig resource.
type Config struct {
	// Dependencies.
	CertSearcher         certs.Interface
	GetClusterConfigFunc func(interface{}) (v1alpha1.ClusterGuestConfig, error)
	K8sClient            kubernetes.Interface
	Logger               micrologger.Logger

	// Settings.
	CertsWatchTimeout time.Duration
	ProjectName       string
}

// StateGetter implements the kubeconfig resource.
type StateGetter struct {
	// Dependencies.
	certsSearcher        certs.Interface
	getClusterConfigFunc func(interface{}) (v1alpha1.ClusterGuestConfig, error)
	k8sClient            kubernetes.Interface
	logger               micrologger.Logger

	// Settings.
	projectName string
}

// New creates a new configured index resource.
func New(config Config) (*StateGetter, error) {
	// Dependencies.
	if config.CertSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertSearcher must not be empty", config)
	}
	if config.GetClusterConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.TransformFunc must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Settings
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName not be empty", config)
	}

	r := &StateGetter{
		// Dependencies.
		certsSearcher:        config.CertSearcher,
		k8sClient:            config.K8sClient,
		logger:               config.Logger,
		getClusterConfigFunc: config.GetClusterConfigFunc,

		// Settings
		projectName: config.ProjectName,
	}

	return r, nil
}

func toSecret(v interface{}) (*corev1.Secret, error) {
	if v == nil {
		return nil, nil
	}

	secret, ok := v.(*corev1.Secret)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", secret, v)
	}

	return secret, nil
}
