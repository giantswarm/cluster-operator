package controller

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
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
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/controller/resource/app"
	"github.com/giantswarm/cluster-operator/service/controller/resource/basedomain"
	"github.com/giantswarm/cluster-operator/service/controller/resource/certconfig"
	"github.com/giantswarm/cluster-operator/service/controller/resource/cleanupmachinedeployments"
	"github.com/giantswarm/cluster-operator/service/controller/resource/clusterconfigmap"
	"github.com/giantswarm/cluster-operator/service/controller/resource/clusterid"
	"github.com/giantswarm/cluster-operator/service/controller/resource/clusterstatus"
	"github.com/giantswarm/cluster-operator/service/controller/resource/cpnamespace"
	"github.com/giantswarm/cluster-operator/service/controller/resource/encryptionkey"
	"github.com/giantswarm/cluster-operator/service/controller/resource/kubeconfig"
	"github.com/giantswarm/cluster-operator/service/controller/resource/operatorversions"
	"github.com/giantswarm/cluster-operator/service/controller/resource/tenantclients"
	"github.com/giantswarm/cluster-operator/service/controller/resource/updatemachinedeployments"
	"github.com/giantswarm/cluster-operator/service/controller/resource/workercount"
)

// clusterResourceSetConfig contains necessary dependencies and settings for
// Cluster API's Cluster controller ResourceSet configuration.
type clusterResourceSetConfig struct {
	CertsSearcher certs.Interface
	ClusterClient *clusterclient.Client
	FileSystem    afero.Fs
	K8sClient     k8sclient.Interface
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

// newClusterResourceSet returns a configured Cluster API's Cluster controller
// ResourceSet.
func newClusterResourceSet(config clusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var appGetter appresource.StateGetter
	{
		c := app.Config{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

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

	var baseDomainResource resource.Interface
	{
		c := basedomain.Config{
			Logger:        config.Logger,
			ToClusterFunc: toClusterFunc,
		}

		baseDomainResource, err = basedomain.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certConfigResource resource.Interface
	{
		c := certconfig.Config{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,

			APIIP:    config.APIIP,
			CertTTL:  config.CertTTL,
			Provider: config.Provider,
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

	var cleanupMachineDeployments resource.Interface
	{
		c := cleanupmachinedeployments.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		cleanupMachineDeployments, err = cleanupmachinedeployments.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConfigMapGetter configmapresource.StateGetter
	{
		c := clusterconfigmap.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			DNSIP: config.DNSIP,
		}

		clusterConfigMapGetter, err = clusterconfigmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConfigMapResource resource.Interface
	{
		c := configmapresource.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			Name:        clusterconfigmap.Name,
			StateGetter: clusterConfigMapGetter,
		}

		ops, err := configmapresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterConfigMapResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterIDResource resource.Interface
	{
		c := clusterid.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewCommonClusterObject: newCommonClusterObjectFunc(config.Provider),
		}

		clusterIDResource, err = clusterid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterStatusResource resource.Interface
	{
		c := clusterstatus.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewCommonClusterObject: newCommonClusterObjectFunc(config.Provider),
			Provider:               config.Provider,
		}

		clusterStatusResource, err = clusterstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cpNamespaceResource resource.Interface
	{
		c := cpnamespace.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		ops, err := cpnamespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		cpNamespaceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptionKeyGetter secretresource.StateGetter
	{
		c := encryptionkey.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		encryptionKeyGetter, err = encryptionkey.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptionKeyResource resource.Interface
	{
		c := secretresource.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			Name:        encryptionkey.Name,
			StateGetter: encryptionKeyGetter,
		}

		ops, err := secretresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		encryptionKeyResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kubeConfigGetter secretresource.StateGetter
	{
		c := kubeconfig.Config{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
		}

		kubeConfigGetter, err = kubeconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kubeConfigResource resource.Interface
	{
		c := secretresource.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			Name:        kubeconfig.Name,
			StateGetter: kubeConfigGetter,
		}

		ops, err := secretresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		kubeConfigResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorVersionsResource resource.Interface
	{
		c := operatorversions.Config{
			ClusterClient: config.ClusterClient,
			Logger:        config.Logger,
		}

		operatorVersionsResource, err = operatorversions.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClientsResource resource.Interface
	{
		c := tenantclients.Config{
			Logger:        config.Logger,
			Tenant:        config.Tenant,
			ToClusterFunc: toClusterFunc,
		}

		tenantClientsResource, err = tenantclients.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var updateMachineDeployments resource.Interface
	{
		c := updatemachinedeployments.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			Provider: config.Provider,
		}

		updateMachineDeployments, err = updatemachinedeployments.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var workerCountResource resource.Interface
	{
		c := workercount.Config{
			Logger: config.Logger,

			ToClusterFunc: toClusterFunc,
		}

		workerCountResource, err = workercount.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// Following resources manage controller context information.
		baseDomainResource,
		operatorVersionsResource,
		tenantClientsResource,
		workerCountResource,

		// Following resources manage CR status information.
		clusterIDResource,
		clusterStatusResource,

		// Following resources manage resources in the control plane.
		cpNamespaceResource,
		encryptionKeyResource,
		certConfigResource,
		clusterConfigMapResource,
		kubeConfigResource,
		appResource,
		updateMachineDeployments,

		// Following resources manage tenant cluster deletion events.
		cleanupMachineDeployments,
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
		cr, err := key.ToCluster(obj)
		if err != nil {
			return false
		}

		if key.OperatorVersion(&cr) == project.BundleVersion() {
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

func newCommonClusterObjectFunc(provider string) func() infrastructurev1alpha2.CommonClusterObject {
	switch provider {
	case "aws":
		return func() infrastructurev1alpha2.CommonClusterObject {
			return new(infrastructurev1alpha2.AWSCluster)
		}

	default:
		panic(fmt.Sprintf("No support for provider %s", provider))
	}
}

func toClusterFunc(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return apiv1alpha2.Cluster{}, microerror.Mask(err)
	}

	return cr, nil
}

func toCRUDResource(logger micrologger.Logger, v crud.Interface) (*crud.Resource, error) {
	c := crud.ResourceConfig{
		CRUD:   v,
		Logger: logger,
	}

	r, err := crud.NewResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
