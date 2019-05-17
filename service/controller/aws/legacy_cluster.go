package aws

import (
	"time"

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
	v10 "github.com/giantswarm/cluster-operator/service/controller/aws/v10"
	v11 "github.com/giantswarm/cluster-operator/service/controller/aws/v11"
	v12 "github.com/giantswarm/cluster-operator/service/controller/aws/v12"
	v13 "github.com/giantswarm/cluster-operator/service/controller/aws/v13"
	v14 "github.com/giantswarm/cluster-operator/service/controller/aws/v14"
	v14patch1 "github.com/giantswarm/cluster-operator/service/controller/aws/v14patch1"
	v15 "github.com/giantswarm/cluster-operator/service/controller/aws/v15"
	v16 "github.com/giantswarm/cluster-operator/service/controller/aws/v16"
)

// LegacyClusterConfig contains necessary dependencies and settings for
// AWSClusterConfig CRD controller implementation.
type LegacyClusterConfig struct {
	ApprClient        *apprclient.Client
	BaseClusterConfig *cluster.Config
	CertSearcher      certs.Interface
	Fs                afero.Fs
	G8sClient         versioned.Interface
	K8sClient         kubernetes.Interface
	K8sExtClient      apiextensionsclient.Interface
	Logger            micrologger.Logger

	CalicoAddress      string
	CalicoPrefixLength string
	ClusterIPRange     string
	ProjectName        string
	RegistryDomain     string
	ResourceNamespace  string
}

type LegacyCluster struct {
	*controller.Controller
}

// NewLegacyCluster returns a configured AWSClusterConfig controller implementation.
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
			Logger: config.Logger,
			// ResyncPeriod is 1 minute because some resources access guest
			// clusters. So we need to wait until they become available. When
			// a guest cluster is not available we cancel the reconciliation.
			ResyncPeriod: 1 * time.Minute,
			Watcher:      config.G8sClient.CoreV1alpha1().AWSClusterConfigs(""),
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v10ResourceSet *controller.ResourceSet
	{
		c := v10.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
		}

		v10ResourceSet, err = v10.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v11ResourceSet *controller.ResourceSet
	{
		c := v11.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
		}

		v11ResourceSet, err = v11.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v12ResourceSet *controller.ResourceSet
	{
		c := v12.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
		}

		v12ResourceSet, err = v12.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v13ResourceSet *controller.ResourceSet
	{
		c := v13.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
		}

		v13ResourceSet, err = v13.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v14ResourceSet *controller.ResourceSet
	{
		c := v14.ResourceSetConfig{
			ApprClient:        config.ApprClient,
			BaseClusterConfig: config.BaseClusterConfig,
			CertSearcher:      config.CertSearcher,
			Fs:                config.Fs,
			G8sClient:         config.G8sClient,
			K8sClient:         config.K8sClient,
			Logger:            config.Logger,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
			ResourceNamespace:  config.ResourceNamespace,
		}

		v14ResourceSet, err = v14.NewResourceSet(c)
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

	var clusterController *controller.Controller
	{
		c := controller.Config{
			CRD:       v1alpha1.NewAWSClusterConfigCRD(),
			CRDClient: crdClient,
			Informer:  newInformer,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				v10ResourceSet,
				v11ResourceSet,
				v12ResourceSet,
				v13ResourceSet,
				v14ResourceSet,
				v14patch1ResourceSet,
				v15ResourceSet,
				v16ResourceSet,
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
