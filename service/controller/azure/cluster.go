package azure

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/spf13/afero"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v1"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v2"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v3"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v4"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v5"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v6"
)

// ClusterConfig contains necessary dependencies and settings for
// AzureClusterConfig CRD controller implementation.
type ClusterConfig struct {
	ApprClient        *apprclient.Client
	BaseClusterConfig *cluster.Config
	CertSearcher      certs.Interface
	Fs                afero.Fs
	G8sClient         versioned.Interface
	K8sClient         kubernetes.Interface
	K8sExtClient      apiextensionsclient.Interface
	Logger            micrologger.Logger

	ProjectName    string
	RegistryDomain string
}

type Cluster struct {
	*controller.Controller
}

// NewCluster returns a configured AzureClusterConfig controller implementation.
func NewCluster(config ClusterConfig) (*Cluster, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}

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
			Watcher: config.G8sClient.CoreV1alpha1().AzureClusterConfigs(""),
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v1ResourceSet *controller.ResourceSet
	{
		c := v1.ResourceSetConfig{
			K8sClient:   config.K8sClient,
			Logger:      config.Logger,
			ProjectName: config.ProjectName,
		}

		v1ResourceSet, err = v1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v2ResourceSet *controller.ResourceSet
	{
		c := v2.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			ProjectName:       config.ProjectName,
		}

		v2ResourceSet, err = v2.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v3ResourceSet *controller.ResourceSet
	{
		c := v3.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			ProjectName:       config.ProjectName,
		}

		v3ResourceSet, err = v3.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v4ResourceSet *controller.ResourceSet
	{
		c := v4.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			ProjectName:       config.ProjectName,
		}

		v4ResourceSet, err = v4.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v5ResourceSet *controller.ResourceSet
	{
		c := v5.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			ProjectName:       config.ProjectName,
			RegistryDomain:    config.RegistryDomain,
		}

		v5ResourceSet, err = v5.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v6ResourceSet *controller.ResourceSet
	{
		c := v6.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			ProjectName:       config.ProjectName,
			RegistryDomain:    config.RegistryDomain,
		}

		v6ResourceSet, err = v6.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			CRD:       v1alpha1.NewAzureClusterConfigCRD(),
			CRDClient: crdClient,
			Informer:  newInformer,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				v1ResourceSet,
				v2ResourceSet,
				v3ResourceSet,
				v4ResourceSet,
				v5ResourceSet,
				v6ResourceSet,
			},
			RESTClient: config.G8sClient.CoreV1alpha1().RESTClient(),

			Name: config.ProjectName,
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
