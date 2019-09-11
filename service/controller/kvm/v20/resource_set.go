package v20

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/appresource"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/k8s/configmapresource"
	"github.com/giantswarm/operatorkit/resource/k8s/secretresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
	chartconfigservice "github.com/giantswarm/cluster-operator/pkg/v20/chartconfig"
	configmapservice "github.com/giantswarm/cluster-operator/pkg/v20/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v20/resource/app"
	"github.com/giantswarm/cluster-operator/pkg/v20/resource/appmigration"
	"github.com/giantswarm/cluster-operator/pkg/v20/resource/certconfig"
	"github.com/giantswarm/cluster-operator/pkg/v20/resource/clusterconfigmap"
	"github.com/giantswarm/cluster-operator/pkg/v20/resource/encryptionkey"
	"github.com/giantswarm/cluster-operator/pkg/v20/resource/kubeconfig"
	"github.com/giantswarm/cluster-operator/service/controller/kvm/v20/key"
	"github.com/giantswarm/cluster-operator/service/controller/kvm/v20/resource/chartconfig"
	"github.com/giantswarm/cluster-operator/service/controller/kvm/v20/resource/configmap"
)

// ResourceSetConfig contains necessary dependencies and settings for
// KVMClusterConfig controller ResourceSet configuration.
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

// NewResourceSet returns a configured KVMClusterConfig controller ResourceSet.
func NewResourceSet(config ResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.ProjectName must not be empty")
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
			Provider:                 label.ProviderKVM,
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

	var chartConfigService chartconfigservice.Interface
	{
		c := chartconfigservice.Config{
			Logger: config.Logger,
			Tenant: config.Tenant,

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

	resources := []resource.Interface{
		// Put encryptionKeyResource first because it executes faster than
		// certConfigResource and could introduce dependency during cluster
		// creation.
		encryptionKeyResource,
		certConfigResource,
		clusterConfigMapResource,
		kubeConfigResource,
		appMigrationResource,
		appResource,
		appMigrationResource,

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
		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		kvmClusterConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(kvmClusterConfig) == VersionBundle().Version {
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
	kvmClusterConfig, err := key.ToCustomObject(obj)
	if err != nil {
		return v1alpha1.ClusterGuestConfig{}, microerror.Mask(err)
	}

	return key.ToClusterGuestConfig(kvmClusterConfig), nil
}

func toClusterObjectMeta(obj interface{}) (metav1.ObjectMeta, error) {
	kvmClusterConfig, err := key.ToCustomObject(obj)
	if err != nil {
		return metav1.ObjectMeta{}, microerror.Mask(err)
	}

	return kvmClusterConfig.ObjectMeta, nil
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
