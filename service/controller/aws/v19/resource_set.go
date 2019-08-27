package v19

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	appresource "github.com/giantswarm/apiextensions/pkg/resource/app"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
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
	chartconfigservice "github.com/giantswarm/cluster-operator/pkg/v19/chartconfig"
	configmapservice "github.com/giantswarm/cluster-operator/pkg/v19/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v19/resource/app"
	"github.com/giantswarm/cluster-operator/pkg/v19/resource/certconfig"
	"github.com/giantswarm/cluster-operator/pkg/v19/resource/chartoperator"
	"github.com/giantswarm/cluster-operator/pkg/v19/resource/clusterconfigmap"
	"github.com/giantswarm/cluster-operator/pkg/v19/resource/encryptionkey"
	"github.com/giantswarm/cluster-operator/pkg/v19/resource/kubeconfig"
	"github.com/giantswarm/cluster-operator/pkg/v19/resource/namespace"
	"github.com/giantswarm/cluster-operator/pkg/v19/resource/tiller"
	"github.com/giantswarm/cluster-operator/service/controller/aws/v19/key"
	"github.com/giantswarm/cluster-operator/service/controller/aws/v19/resource/chartconfig"
	"github.com/giantswarm/cluster-operator/service/controller/aws/v19/resource/configmap"
)

// ResourceSetConfig contains necessary dependencies and settings for
// AWSClusterConfig controller ResourceSet configuration.
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

// NewResourceSet returns a configured AWSClusterConfig controller ResourceSet.
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
	if config.ResourceNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ResourceNamespace must not be empty", config)
	}

	var appGetter appresource.StateGetter
	{
		c := app.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			Provider: config.Provider,
		}

		appGetter, err = app.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appResource controller.Resource
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

	var certConfigResource controller.Resource
	{
		c := certconfig.Config{
			BaseClusterConfig:        *config.BaseClusterConfig,
			G8sClient:                config.G8sClient,
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,
			Provider:                 label.ProviderAWS,
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

	var namespaceResource controller.Resource
	{
		c := namespace.Config{
			BaseClusterConfig:        *config.BaseClusterConfig,
			Logger:                   config.Logger,
			Tenant:                   config.Tenant,
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

	var chartOperatorResource controller.Resource
	{
		c := chartoperator.Config{
			ApprClient:               config.ApprClient,
			BaseClusterConfig:        *config.BaseClusterConfig,
			ClusterIPRange:           config.ClusterIPRange,
			Fs:                       config.Fs,
			G8sClient:                config.G8sClient,
			K8sClient:                config.K8sClient,
			Logger:                   config.Logger,
			ProjectName:              config.ProjectName,
			RegistryDomain:           config.RegistryDomain,
			Tenant:                   config.Tenant,
			ToClusterGuestConfigFunc: toClusterGuestConfig,
			ToClusterObjectMetaFunc:  toClusterObjectMeta,
		}

		ops, err := chartoperator.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		chartOperatorResource, err = toCRUDResource(config.Logger, ops)
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
			Tenant: config.Tenant,
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

	var clusterConfigMapResource controller.Resource
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

	var kubeConfigResource controller.Resource
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

	var tillerResource controller.Resource
	{
		c := tiller.Config{
			BaseClusterConfig:        *config.BaseClusterConfig,
			Logger:                   config.Logger,
			Tenant:                   config.Tenant,
			ToClusterGuestConfigFunc: toClusterGuestConfig,
			ToClusterObjectMetaFunc:  toClusterObjectMeta,
		}

		tillerResource, err = tiller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		// Put encryptionKeyResource first because it executes faster than
		// certConfigResource and could introduce dependency during cluster
		// creation.
		encryptionKeyResource,
		certConfigResource,
		clusterConfigMapResource,
		kubeConfigResource,
		appResource,

		// Following resources manage resources in tenant clusters so they
		// should be executed last.
		namespaceResource,
		tillerResource,
		chartOperatorResource,
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
		cr, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.OperatorVersion(&cr) == VersionBundle().Version {
			return true
		}

		if key.VersionBundleVersion(cr) == VersionBundle().Version {
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
	awsClusterConfig, err := key.ToCustomObject(obj)
	if err != nil {
		return v1alpha1.ClusterGuestConfig{}, microerror.Mask(err)
	}

	return key.ClusterGuestConfig(awsClusterConfig), nil
}

func toClusterObjectMeta(obj interface{}) (metav1.ObjectMeta, error) {
	awsClusterConfig, err := key.ToCustomObject(obj)
	if err != nil {
		return metav1.ObjectMeta{}, microerror.Mask(err)
	}

	return awsClusterConfig.ObjectMeta, nil
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
