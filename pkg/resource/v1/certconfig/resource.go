package certconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "certconfigv1"

	// Maximum number of CertConfigs returned in one List() call to K8s API
	listCertConfigLimit = 25
)

type certificate struct {
	name              certs.Cert
	certConfigFactory func() (*v1alpha1.CertConfig, error)
}

// This is a list of certificates managed by this resource.
var managedCertificates = []certificate{
	{
		name: certs.APICert,
	},
	{
		name: certs.CalicoCert,
	},
	{
		name: certs.EtcdCert,
	},
	{
		name: certs.FlanneldCert,
	},
	{
		name: certs.NodeOperatorCert,
	},
	{
		name: certs.PrometheusCert,
	},
	{
		name: certs.ServiceAccountCert,
	},
	{
		name: certs.WorkerCert,
	},
}

// Config represents the configuration used to create a new cloud config resource.
type Config struct {
	G8sClient                versioned.Interface
	K8sClient                kubernetes.Interface
	Logger                   micrologger.Logger
	ProjectName              string
	ToClusterGuestConfigFunc func(obj interface{}) (*v1alpha1.ClusterGuestConfig, error)
}

// Resource implements the cloud config resource.
type Resource struct {
	g8sClient                versioned.Interface
	k8sClient                kubernetes.Interface
	logger                   micrologger.Logger
	projectName              string
	toClusterGuestConfigFunc func(obj interface{}) (*v1alpha1.ClusterGuestConfig, error)
}

// New creates a new configured cloud config resource.
func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.G8sClient must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.ProjectName must not be empty")
	}
	if config.ToClusterGuestConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.ToClusterGuestConfigFunc must not be empty")
	}

	newService := &Resource{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger: config.Logger.With(
			"resource", Name,
		),
		projectName:              config.ProjectName,
		toClusterGuestConfigFunc: config.ToClusterGuestConfigFunc,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}
