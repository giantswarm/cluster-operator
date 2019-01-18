package v7patch1

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
	chartconfigservice "github.com/giantswarm/cluster-operator/pkg/v7patch1/chartconfig"
	configmapservice "github.com/giantswarm/cluster-operator/pkg/v7patch1/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v7patch1/resource/certconfig"
	"github.com/giantswarm/cluster-operator/pkg/v7patch1/resource/chart"
	"github.com/giantswarm/cluster-operator/pkg/v7patch1/resource/clustercr"
	"github.com/giantswarm/cluster-operator/pkg/v7patch1/resource/encryptionkey"
	"github.com/giantswarm/cluster-operator/pkg/v7patch1/resource/namespace"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v7patch1/key"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v7patch1/resource/azureconfig"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v7patch1/resource/chartconfig"
	"github.com/giantswarm/cluster-operator/service/controller/azure/v7patch1/resource/configmap"
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

	CalicoAddress         string
	CalicoPrefixLength    string
	ClusterIPRange        string
	HandledVersionBundles []string
	ProjectName           string
	RegistryDomain        string
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

	var certConfigResource controller.Resource
	{
		c := certconfig.Config{
			BaseClusterConfig:        *config.BaseClusterConfig,
			G8sClient:                config.G8sClient,
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,
			ProjectName:              config.ProjectName,
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

	var encryptionKeyResource controller.Resource
	{
		c := encryptionkey.Config{
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,
			ProjectName:              config.ProjectName,
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

	var azureConfigResource controller.Resource
	{
		c := azureconfig.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		ops, err := azureconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		azureConfigResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClusterService tenantcluster.Interface
	{
		c := tenantcluster.Config{
			CertsSearcher: config.CertSearcher,
			Logger:        config.Logger,

			CertID: certs.ClusterOperatorAPICert,
		}

		tenantClusterService, err = tenantcluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource controller.Resource
	{
		c := namespace.Config{
			BaseClusterConfig:        *config.BaseClusterConfig,
			Logger:                   config.Logger,
			ProjectName:              config.ProjectName,
			Tenant:                   tenantClusterService,
			ToClusterGuestConfigFunc: toClusterGuestConfig,
			ToClusterObjectMetaFunc:  toClusterObjectMeta,
		}

		ops, err := namespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		namespaceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartResource controller.Resource
	{
		c := chart.Config{
			ApprClient:               config.ApprClient,
			BaseClusterConfig:        *config.BaseClusterConfig,
			ClusterIPRange:           config.ClusterIPRange,
			Fs:                       config.Fs,
			G8sClient:                config.G8sClient,
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,
			ProjectName:              config.ProjectName,
			RegistryDomain:           config.RegistryDomain,
			Tenant:                   tenantClusterService,
			ToClusterGuestConfigFunc: toClusterGuestConfig,
		}

		ops, err := chart.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		chartResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapService configmapservice.Interface
	{
		c := configmapservice.Config{
			Logger: config.Logger,
			Tenant: tenantClusterService,

			ProjectName: config.ProjectName,
		}

		configMapService, err = configmapservice.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource controller.Resource
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
			Tenant: tenantClusterService,

			ProjectName: config.ProjectName,
		}

		chartConfigService, err = chartconfigservice.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartConfigResource controller.Resource
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

	var clusterCRResource controller.Resource
	{
		c := clustercr.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		clusterCRResource, err = clustercr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		clusterCRResource,
		// Put encryptionKeyResource first because it executes faster than
		// azureConfigResource and could introduce dependency during cluster
		// creation.
		encryptionKeyResource,
		certConfigResource,
		azureConfigResource,
		// namespace, chart, configmap and chartconfig resources manage resources
		// in guest clusters so they should be executed last.
		namespaceResource,
		chartResource,
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

func toClusterGuestConfig(obj interface{}) (v1alpha1.ClusterGuestConfig, error) {
	azureClusterConfig, err := key.ToCustomObject(obj)
	if err != nil {
		return v1alpha1.ClusterGuestConfig{}, microerror.Mask(err)
	}

	return key.ClusterGuestConfig(azureClusterConfig), nil
}

func toClusterObjectMeta(obj interface{}) (apismetav1.ObjectMeta, error) {
	azureClusterConfig, err := key.ToCustomObject(obj)
	if err != nil {
		return apismetav1.ObjectMeta{}, microerror.Mask(err)
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
