package kubeconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/cluster-operator/pkg/v13/chartconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/v13/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v13/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v13/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v13/key"
)

// Config represents the configuration used to create a new index resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Settings.
	ProjectName       string
	ResourceName      string
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
	resourceName      string
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
	if config.ResourceName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ResourceName must not be empty", config)
	}
	if config.ResourceNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ResourceNamespace must not be empty", config)
	}

	var cert certs.Interface
	{
		var err error
		cc := certs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
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
		resourceName:      config.ResourceName,
		resourceNamespace: config.ResourceNamespace,
	}

	return r, nil
}

func (r *StateGetter) Name() string {
	return r.resourceName
}

func ToCustomObject(v interface{}) (*chartconfig.ClusterConfig, error) {
	var guestConfig v1alpha1.ClusterGuestConfig
	switch object := v.(type) {
	case *v1alpha1.AWSClusterConfig:
		guestConfig = awskey.ClusterGuestConfig(*object)
	case *v1alpha1.AzureClusterConfig:
		guestConfig = azurekey.ClusterGuestConfig(*object)
	case *v1alpha1.KVMClusterConfig:
		guestConfig = kvmkey.ClusterGuestConfig(*object)
	default:
		return nil, microerror.Maskf(invalidConfigError, "cannot identify interface %#v", v)
	}

	apiDomain, err := key.APIDomain(guestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig := chartconfig.ClusterConfig{
		APIDomain:    apiDomain,
		ClusterID:    key.ClusterID(guestConfig),
		Organization: key.ClusterOrganization(guestConfig),
	}
	return &clusterConfig, nil
}

func toSecret(v interface{}) (*v1.Secret, error) {
	x, ok := v.(*v1.Secret)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", x, v)
	}

	return x, nil
}
