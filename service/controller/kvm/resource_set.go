package kvm

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/crud"
	"github.com/giantswarm/operatorkit/resource/k8s/configmapresource"
	"github.com/giantswarm/operatorkit/resource/k8s/secretresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/resource/appresource"
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/kvm/key"
	"github.com/giantswarm/cluster-operator/service/controller/resource/app"
	"github.com/giantswarm/cluster-operator/service/controller/resource/appmigration"
	"github.com/giantswarm/cluster-operator/service/controller/resource/certconfig"
	"github.com/giantswarm/cluster-operator/service/controller/resource/clusterconfigmap"
	"github.com/giantswarm/cluster-operator/service/controller/resource/configmapmigration"
	"github.com/giantswarm/cluster-operator/service/controller/resource/encryptionkey"
	"github.com/giantswarm/cluster-operator/service/controller/resource/kubeconfig"
	"github.com/giantswarm/cluster-operator/service/controller/resource/tenantclients"
	"github.com/giantswarm/cluster-operator/service/controller/resource/workercount"
	"github.com/giantswarm/cluster-operator/service/internal/cluster"
)

type resourceSetConfig struct {
	ApprClient        *apprclient.Client
	BaseClusterConfig *cluster.Config
	CertSearcher      certs.Interface
	ClusterClient     *clusterclient.Client
	Fs                afero.Fs
	K8sClient         k8sclient.Interface
	Logger            micrologger.Logger
	Tenant            tenantcluster.Interface

	CalicoAddress         string
	CalicoPrefixLength    string
	ClusterIPRange        string
	HandledVersionBundles []string
	ProjectName           string
	Provider              string
	RawAppDefaultConfig   string
	RawAppOverrideConfig  string
	RegistryDomain        string
	ResourceNamespace     string
}

func newResourceSet(config resourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	if config.ClusterClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterClient must not be empty", config)
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

	var appGetter appresource.StateGetter
	{
		c := app.Config{
			ClusterClient:            config.ClusterClient,
			G8sClient:                config.K8sClient.G8sClient(),
			GetClusterConfigFunc:     getClusterConfig,
			GetClusterObjectMetaFunc: getClusterObjectMeta,
			K8sClient:                config.K8sClient.K8sClient(),
			Logger:                   config.Logger,

			Provider:             config.Provider,
			RawAppDefaultConfig:  config.RawAppDefaultConfig,
			RawAppOverrideConfig: config.RawAppOverrideConfig,
		}

		appGetter, err = app.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appResource resource.Interface
	{
		c := appresource.Config{
			G8sClient: config.K8sClient.G8sClient(),
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
			G8sClient:                config.K8sClient.G8sClient(),
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
			G8sClient:                config.K8sClient.G8sClient(),
			K8sClient:                config.K8sClient.K8sClient(),
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
			K8sClient:                config.K8sClient.K8sClient(),
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

	var configMapMigrationResource resource.Interface
	{
		c := configmapmigration.Config{
			GetClusterConfigFunc:     getClusterConfig,
			GetClusterObjectMetaFunc: getClusterObjectMeta,
			K8sClient:                config.K8sClient.K8sClient(),
			Logger:                   config.Logger,

			Provider: config.Provider,
		}

		configMapMigrationResource, err = configmapmigration.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConfigMapResource resource.Interface
	{
		c := clusterconfigmap.Config{
			GetClusterConfigFunc:     getClusterConfig,
			GetClusterObjectMetaFunc: getClusterObjectMeta,
			GetWorkerCountFunc:       getWorkerCount,
			K8sClient:                config.K8sClient.K8sClient(),
			Logger:                   config.Logger,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			Provider:           config.Provider,
		}

		stateGetter, err := clusterconfigmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configOps := configmapresource.Config{
			K8sClient: config.K8sClient.K8sClient(),
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
			K8sClient:                config.K8sClient.K8sClient(),
			Logger:                   config.Logger,
		}

		stateGetter, err := kubeconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configOps := secretresource.Config{
			K8sClient: config.K8sClient.K8sClient(),
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

	var workerCountResource resource.Interface
	{
		c := workercount.Config{
			Logger:              config.Logger,
			ToClusterConfigFunc: getClusterConfig,
		}

		workerCountResource, err = workercount.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// Following resources manage resources controller context information.
		tenantClientsResource,
		workerCountResource,

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
		kvmClusterConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(kvmClusterConfig) == project.BundleVersion() {
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

func getWorkerCount(obj interface{}) (int, error) {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.WorkerCount(cr), nil
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

func toCRUDResource(logger micrologger.Logger, ops crud.Interface) (resource.Interface, error) {
	c := crud.ResourceConfig{
		CRUD:   ops,
		Logger: logger,
	}

	r, err := crud.NewResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
