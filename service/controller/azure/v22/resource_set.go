package v22

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/k8s/configmapresource"
	"github.com/giantswarm/operatorkit/resource/k8s/secretresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/resource/appresource"
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
	chartconfigservice "github.com/giantswarm/cluster-operator/pkg/v22/chartconfig"
	configmapservice "github.com/giantswarm/cluster-operator/pkg/v22/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v22/controllercontext"
	"github.com/giantswarm/cluster-operator/pkg/v22/resource/app"
	"github.com/giantswarm/cluster-operator/pkg/v22/resource/appmigration"
	"github.com/giantswarm/cluster-operator/pkg/v22/resource/certconfig"
	"github.com/giantswarm/cluster-operator/pkg/v22/resource/clusterconfigmap"
	"github.com/giantswarm/cluster-operator/pkg/v22/resource/configmapmigration"
	"github.com/giantswarm/cluster-operator/pkg/v22/resource/encryptionkey"
	"github.com/giantswarm/cluster-operator/pkg/v22/resource/kubeconfig"
	"github.com/giantswarm/cluster-operator/pkg/v22/resource/tenantclients"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v22/key"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v22/resource/chartconfig"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v22/resource/configmap"
)

// ResourceSetConfig contains necessary dependencies and settings for
// AzureClusterConfig controller ResourceSet configuration.
type ResourceSetConfig struct {
	ApprClient        *apprclient.Client
	BaseClusterConfig *cluster.Config
	CertSearcher      certs.Interface
	Fs                afero.Fs
	G8sClient         versioned.Interface
	K8sClient         kubernetes.Interface
	Logger            micrologger.Logger
	Tenant            tenantcluster.Interface

	CalicoAddress         string
	CalicoPrefixLength    string
	ClusterIPRange        string
	HandledVersionBundles []string
	ProjectName           string
	Provider              string
	RegistryDomain        string
	ResourceNamespace     string
}

// NewResourceSet returns a configured AzureClusterConfig controller ResourceSet.
func NewResourceSet(config ResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	var appGetter appresource.StateGetter
	{
		c := app.Config{
			G8sClient:                config.G8sClient,
			GetClusterConfigFunc:     getClusterConfig,
			GetClusterObjectMetaFunc: getClusterObjectMeta,
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,

			Provider: config.Provider,
		}

		appGetter, err = app.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appResource resource.Interface
	{
		c := appresource.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			Name:        app.Name,
			StateGetter: appGetter,
		}

		ops, err := appresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		appResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appMigrationResource resource.Interface
	{
		c := appmigration.Config{
			GetClusterConfigFunc:     getClusterConfig,
			GetClusterObjectMetaFunc: getClusterObjectMeta,
			G8sClient:                config.G8sClient,
			Logger:                   config.Logger,
			Tenant:                   config.Tenant,

			Provider: config.Provider,
		}

		appMigrationResource, err = appmigration.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certConfigResource resource.Interface
	{
		c := certconfig.Config{
			BaseClusterConfig:        *config.BaseClusterConfig,
			G8sClient:                config.G8sClient,
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,
			Provider:                 label.ProviderAzure,
			ToClusterGuestConfigFunc: toClusterGuestConfig,
			ToClusterObjectMetaFunc:  toClusterObjectMeta,
		}

		ops, err := certconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		certConfigResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptionKeyResource resource.Interface
	{
		c := encryptionkey.Config{
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,
			ToClusterGuestConfigFunc: toClusterGuestConfig,
			ToClusterObjectMetaFunc:  toClusterObjectMeta,
		}

		ops, err := encryptionkey.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		encryptionKeyResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapService configmapservice.Interface
	{
		c := configmapservice.Config{
			Logger: config.Logger,
			Tenant: config.Tenant,

			Provider: config.Provider,
		}

		configMapService, err = configmapservice.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource resource.Interface
	{
		c := configmap.Config{
			ConfigMap: configMapService,
			Logger:    config.Logger,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			ProjectName:        config.ProjectName,
			RegistryDomain:     config.RegistryDomain,
		}

		ops, err := configmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMapResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapMigrationResource resource.Interface
	{
		c := configmapmigration.Config{
			GetClusterConfigFunc:     getClusterConfig,
			GetClusterObjectMetaFunc: getClusterObjectMeta,
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,

			Provider: config.Provider,
		}

		configMapMigrationResource, err = configmapmigration.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartConfigService chartconfigservice.Interface
	{
		c := chartconfigservice.Config{
			Logger: config.Logger,

			Provider: config.Provider,
		}

		chartConfigService, err = chartconfigservice.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartConfigResource resource.Interface
	{
		c := chartconfig.Config{
			ChartConfig: chartConfigService,
			Logger:      config.Logger,

			ProjectName: config.ProjectName,
		}

		ops, err := chartconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		chartConfigResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConfigMapResource resource.Interface
	{
		c := clusterconfigmap.Config{
			GetClusterConfigFunc:     getClusterConfig,
			GetClusterObjectMetaFunc: getClusterObjectMeta,
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,

			ClusterIPRange: config.ClusterIPRange,
		}

		stateGetter, err := clusterconfigmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configOps := configmapresource.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			Name:        clusterconfigmap.Name,
			StateGetter: stateGetter,
		}

		ops, err := configmapresource.New(configOps)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterConfigMapResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kubeConfigResource resource.Interface
	{
		c := kubeconfig.Config{
			CertSearcher:             config.CertSearcher,
			GetClusterConfigFunc:     getClusterConfig,
			GetClusterObjectMetaFunc: getClusterObjectMeta,
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,
		}

		stateGetter, err := kubeconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configOps := secretresource.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			Name:        kubeconfig.Name,
			StateGetter: stateGetter,
		}

		ops, err := secretresource.New(configOps)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		kubeConfigResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClientsResource resource.Interface
	{
		c := tenantclients.Config{
			Logger:              config.Logger,
			Tenant:              config.Tenant,
			ToClusterConfigFunc: getClusterConfig,
		}

		tenantClientsResource, err = tenantclients.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// Following resources manage resources controller context information.
		tenantClientsResource,

		// Following resources manage resources in the control plane.
		encryptionKeyResource,
		certConfigResource,
		clusterConfigMapResource,
		kubeConfigResource,

		// Migration resources are for migrating from chartconfig to app CRs.
		appMigrationResource,
		configMapMigrationResource,

		// appResource is executed after migration resources.
		appResource,

		// appResource is executed after migration resources.
		appResource,

		// Following resources manage resources in tenant clusters so they
		// should be executed last
		configMapResource,
		chartConfigResource,
	}

	// Wrap resources with retry and metrics.
	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}
		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		ctx = controllercontext.NewContext(ctx, controllercontext.Context{})
		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		azureClusterConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(azureClusterConfig) == VersionBundle().Version {
			return true
		}

		return false
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}

func getClusterConfig(obj interface{}) (v1alpha1.ClusterGuestConfig, error) {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return v1alpha1.ClusterGuestConfig{}, microerror.Mask(err)
	}

	return key.ClusterGuestConfig(cr), nil
}

func getClusterObjectMeta(obj interface{}) (metav1.ObjectMeta, error) {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return metav1.ObjectMeta{}, microerror.Mask(err)
	}

	return cr.ObjectMeta, nil
}

func toClusterGuestConfig(obj interface{}) (v1alpha1.ClusterGuestConfig, error) {
	azureClusterConfig, err := key.ToCustomObject(obj)
	if err != nil {
		return v1alpha1.ClusterGuestConfig{}, microerror.Mask(err)
	}

	return key.ClusterGuestConfig(azureClusterConfig), nil
}

func toClusterObjectMeta(obj interface{}) (metav1.ObjectMeta, error) {
	azureClusterConfig, err := key.ToCustomObject(obj)
	if err != nil {
		return metav1.ObjectMeta{}, microerror.Mask(err)
	}

	return azureClusterConfig.ObjectMeta, nil
}

func toCRUDResource(logger micrologger.Logger, ops controller.CRUDResourceOps) (*controller.CRUDResource, error) {
	c := controller.CRUDResourceConfig{
		Logger: logger,
		Ops:    ops,
	}

	r, err := controller.NewCRUDResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
