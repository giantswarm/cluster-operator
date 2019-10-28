package kvm

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
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/service/controller/kvm/v14patch1"
	v15 "github.com/giantswarm/cluster-operator/service/controller/kvm/v15"
	v16 "github.com/giantswarm/cluster-operator/service/controller/kvm/v16"
	v17 "github.com/giantswarm/cluster-operator/service/controller/kvm/v17"
	v18 "github.com/giantswarm/cluster-operator/service/controller/kvm/v18"
	v19 "github.com/giantswarm/cluster-operator/service/controller/kvm/v19"
	v20 "github.com/giantswarm/cluster-operator/service/controller/kvm/v20"
	v21 "github.com/giantswarm/cluster-operator/service/controller/kvm/v21"
	v22 "github.com/giantswarm/cluster-operator/service/controller/kvm/v22"
)

// LegacyClusterConfig contains necessary dependencies and settings for
// KVMClusterConfig CRD controller implementation.
type LegacyClusterConfig struct {
	ApprClient        *apprclient.Client
	BaseClusterConfig *cluster.Config
	CertSearcher      certs.Interface
	Fs                afero.Fs
	G8sClient         versioned.Interface
	K8sClient         kubernetes.Interface
	K8sExtClient      apiextensionsclient.Interface
	Logger            micrologger.Logger
	Tenant            tenantcluster.Interface

	CalicoAddress      string
	CalicoPrefixLength string
	ClusterIPRange     string
	ProjectName        string
	Provider           string
	RegistryDomain     string
	ResourceNamespace  string
}

type LegacyCluster struct {
	*controller.Controller
}

// NewLegacyCluster returns a configured KVMClusterConfig controller implementation.
func NewLegacyCluster(config LegacyClusterConfig) (*LegacyCluster, error) {
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
			Watcher: config.G8sClient.CoreV1alpha1().KVMClusterConfigs(""),
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v14patch1ResourceSet *controller.ResourceSet
	{
		c := v14patch1.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
			ResourceNamespace:  config.ResourceNamespace,
		}

		v14patch1ResourceSet, err = v14patch1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v15ResourceSet *controller.ResourceSet
	{
		c := v15.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
			ResourceNamespace:  config.ResourceNamespace,
		}

		v15ResourceSet, err = v15.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v16ResourceSet *controller.ResourceSet
	{
		c := v16.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
			ResourceNamespace:  config.ResourceNamespace,
		}

		v16ResourceSet, err = v16.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v17ResourceSet *controller.ResourceSet
	{
		c := v17.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
			ResourceNamespace:  config.ResourceNamespace,
		}

		v17ResourceSet, err = v17.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v18ResourceSet *controller.ResourceSet
	{
		c := v18.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
			ResourceNamespace:  config.ResourceNamespace,
		}

		v18ResourceSet, err = v18.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v19ResourceSet *controller.ResourceSet
	{
		c := v19.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			Provider:           config.Provider,
			RegistryDomain:     config.RegistryDomain,
			ResourceNamespace:  config.ResourceNamespace,
		}

		v19ResourceSet, err = v19.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v20ResourceSet *controller.ResourceSet
	{
		c := v20.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			Provider:           config.Provider,
			RegistryDomain:     config.RegistryDomain,
			ResourceNamespace:  config.ResourceNamespace,
		}

		v20ResourceSet, err = v20.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v21ResourceSet *controller.ResourceSet
	{
		c := v21.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			Provider:           config.Provider,
			RegistryDomain:     config.RegistryDomain,
			ResourceNamespace:  config.ResourceNamespace,
		}

		v21ResourceSet, err = v21.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v22ResourceSet *controller.ResourceSet
	{
		c := v22.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,
			Tenant:            config.Tenant,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			Provider:           config.Provider,
			RegistryDomain:     config.RegistryDomain,
			ResourceNamespace:  config.ResourceNamespace,
		}

		v22ResourceSet, err = v22.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			CRD:       v1alpha1.NewKVMClusterConfigCRD(),
			CRDClient: crdClient,
			Informer:  newInformer,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				v14patch1ResourceSet,
				v15ResourceSet,
				v16ResourceSet,
				v17ResourceSet,
				v18ResourceSet,
				v19ResourceSet,
				v20ResourceSet,
				v21ResourceSet,
				v22ResourceSet,
			},
			RESTClient: config.G8sClient.CoreV1alpha1().RESTClient(),

			Name: config.ProjectName,
		}

		clusterController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &LegacyCluster{
		Controller: clusterController,
	}

	return c, nil
}
