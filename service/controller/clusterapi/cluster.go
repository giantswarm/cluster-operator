package clusterapi

import (
	clusterv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/pkg/project"
	v19 "github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19"
	v20 "github.com/giantswarm/cluster-operator/service/controller/clusterapi/v20"
	v21 "github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21"
)

// ClusterConfig contains necessary dependencies and settings for
// Cluster API's Cluster CRD controller implementation.
type ClusterConfig struct {
	ApprClient    *apprclient.Client
	CertsSearcher certs.Interface
	ClusterClient *clusterclient.Client
	CMAClient     clientset.Interface
	FileSystem    afero.Fs
	G8sClient     versioned.Interface
	K8sClient     kubernetes.Interface
	K8sExtClient  apiextensionsclient.Interface
	Logger        micrologger.Logger
	Tenant        tenantcluster.Interface

	APIIP              string
	CalicoAddress      string
	CalicoPrefixLength string
	CertTTL            string
	ClusterIPRange     string
	DNSIP              string
	Provider           string
	RegistryDomain     string
}

type Cluster struct {
	*controller.Controller
}

// NewCluster returns a configured AWSClusterConfig controller implementation.
func NewCluster(config ClusterConfig) (*Cluster, error) {
	var err error

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.Config{
			K8sExtClient: config.K8sExtClient,
			Logger:       config.Logger,
		}

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Logger:  config.Logger,
			Watcher: config.CMAClient.ClusterV1alpha1().Clusters(corev1.NamespaceAll),

			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV19 *controller.ResourceSet
	{
		c := v19.ClusterResourceSetConfig{
			ApprClient:    config.ApprClient,
			CertsSearcher: config.CertsSearcher,
			ClusterClient: config.ClusterClient,
			CMAClient:     config.CMAClient,
			FileSystem:    config.FileSystem,
			G8sClient:     config.G8sClient,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			Tenant:        config.Tenant,

			APIIP:              config.APIIP,
			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			CertTTL:            config.CertTTL,
			ClusterIPRange:     config.ClusterIPRange,
			DNSIP:              config.DNSIP,
			Provider:           config.Provider,
			RegistryDomain:     config.RegistryDomain,
		}

		resourceSetV19, err = v19.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV20 *controller.ResourceSet
	{
		c := v20.ClusterResourceSetConfig{
			ApprClient:    config.ApprClient,
			CertsSearcher: config.CertsSearcher,
			ClusterClient: config.ClusterClient,
			CMAClient:     config.CMAClient,
			FileSystem:    config.FileSystem,
			G8sClient:     config.G8sClient,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			Tenant:        config.Tenant,

			APIIP:              config.APIIP,
			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			CertTTL:            config.CertTTL,
			ClusterIPRange:     config.ClusterIPRange,
			DNSIP:              config.DNSIP,
			Provider:           config.Provider,
			RegistryDomain:     config.RegistryDomain,
		}

		resourceSetV20, err = v20.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV21 *controller.ResourceSet
	{
		c := v21.ClusterResourceSetConfig{
			ApprClient:    config.ApprClient,
			CertsSearcher: config.CertsSearcher,
			ClusterClient: config.ClusterClient,
			CMAClient:     config.CMAClient,
			FileSystem:    config.FileSystem,
			G8sClient:     config.G8sClient,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
			Tenant:        config.Tenant,

			APIIP:              config.APIIP,
			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			CertTTL:            config.CertTTL,
			ClusterIPRange:     config.ClusterIPRange,
			DNSIP:              config.DNSIP,
			Provider:           config.Provider,
			RegistryDomain:     config.RegistryDomain,
		}

		resourceSetV21, err = v21.NewClusterResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			CRD:       clusterv1alpha1.NewClusterCRD(),
			CRDClient: crdClient,
			Informer:  newInformer,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSetV19,
				resourceSetV20,
				resourceSetV21,
			},
			RESTClient: config.CMAClient.ClusterV1alpha1().RESTClient(),

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/cluster-operator-cluster-controller.
			Name: project.Name() + "-cluster-controller",
		}

		clusterController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Cluster{
		Controller: clusterController,
	}

	return c, nil
}
