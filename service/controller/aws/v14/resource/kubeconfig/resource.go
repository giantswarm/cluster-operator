package kubeconfig

import (
	"time"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "kubeconfigv14"
	// DefaultWatchTimeout is the time to wait on watches against the Kubernetes
	// API before giving up and throwing an error.
	DefaultWatchTimeout = 90 * time.Second
)

// Config represents the configuration used to create a new kubeconfig resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Settings.
	CertsWatchTimeout time.Duration
	ProjectName       string
	ResourceNamespace string
}

// StateGetter implements the kubeconfig resource.
type StateGetter struct {
	// Dependencies.
	certsSearcher certs.Interface
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger

	// Settings.
	projectName       string
	resourceNamespace string
}

// New creates a new configured index resource.
func New(config Config) (*StateGetter, error) {
	// Dependencies.
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
	if config.ResourceNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ResourceNamespace not be empty", config)
	}
	if config.CertsWatchTimeout == 0 {
		config.CertsWatchTimeout = DefaultWatchTimeout
	}

	var cert certs.Interface
	{
		var err error
		cc := certs.Config{
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			WatchTimeout: config.CertsWatchTimeout,
		}
		cert, err = certs.NewSearcher(cc)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	r := &StateGetter{
		// Dependencies.
		certsSearcher: cert,
		k8sClient:     config.K8sClient,
		logger:        config.Logger,

		// Settings
		projectName:       config.ProjectName,
		resourceNamespace: config.ResourceNamespace,
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
