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
	"github.com/spf13/afero"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	v10 "github.com/giantswarm/cluster-operator/service/controller/kvm/v10"
	v11 "github.com/giantswarm/cluster-operator/service/controller/kvm/v11"
	v12 "github.com/giantswarm/cluster-operator/service/controller/kvm/v12"
	v13 "github.com/giantswarm/cluster-operator/service/controller/kvm/v13"
	v6 "github.com/giantswarm/cluster-operator/service/controller/kvm/v6"
	"github.com/giantswarm/cluster-operator/service/controller/kvm/v6patch1"
	v7 "github.com/giantswarm/cluster-operator/service/controller/kvm/v7"
	"github.com/giantswarm/cluster-operator/service/controller/kvm/v7patch1"
	v8 "github.com/giantswarm/cluster-operator/service/controller/kvm/v8"
	v9 "github.com/giantswarm/cluster-operator/service/controller/kvm/v9"
)

// ClusterConfig contains necessary dependencies and settings for
// KVMClusterConfig CRD controller implementation.
type ClusterConfig struct {
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
}

type Cluster struct {
	*controller.Controller
}

// NewCluster returns a configured KVMClusterConfig controller implementation.
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
			Watcher: config.G8sClient.CoreV1alpha1().KVMClusterConfigs(""),
		}

		newInformer, err = informer.New(c)
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

	var v6patch1ResourceSet *controller.ResourceSet
	{
		c := v6patch1.ResourceSetConfig{
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

		v6patch1ResourceSet, err = v6patch1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v7ResourceSet *controller.ResourceSet
	{
		c := v7.ResourceSetConfig{
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

		v7ResourceSet, err = v7.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v7patch1ResourceSet *controller.ResourceSet
	{
		c := v7patch1.ResourceSetConfig{
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

		v7patch1ResourceSet, err = v7patch1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v8ResourceSet *controller.ResourceSet
	{
		c := v8.ResourceSetConfig{
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

		v8ResourceSet, err = v8.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v9ResourceSet *controller.ResourceSet
	{
		c := v9.ResourceSetConfig{
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

		v9ResourceSet, err = v9.NewResourceSet(c)
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

	var clusterController *controller.Controller
	{
		c := controller.Config{
			CRD:       v1alpha1.NewKVMClusterConfigCRD(),
			CRDClient: crdClient,
			Informer:  newInformer,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				v6ResourceSet,
				v6patch1ResourceSet,
				v7ResourceSet,
				v7patch1ResourceSet,
				v8ResourceSet,
				v9ResourceSet,
				v10ResourceSet,
				v11ResourceSet,
				v12ResourceSet,
				v13ResourceSet,
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
