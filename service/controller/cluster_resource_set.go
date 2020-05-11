package controller

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs/v2/pkg/certs"
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
	"github.com/giantswarm/tenantcluster/v2/pkg/tenantcluster"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/internal/hamaster"
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
	"github.com/giantswarm/cluster-operator/service/controller/resource/keepforinfrarefs"
	"github.com/giantswarm/cluster-operator/service/controller/resource/kubeconfig"
	"github.com/giantswarm/cluster-operator/service/controller/resource/releaseversions"
	"github.com/giantswarm/cluster-operator/service/controller/resource/statuscondition"
	"github.com/giantswarm/cluster-operator/service/controller/resource/tenantclients"
	"github.com/giantswarm/cluster-operator/service/controller/resource/updateg8scontrolplanes"
	"github.com/giantswarm/cluster-operator/service/controller/resource/updateinfrarefs"
	"github.com/giantswarm/cluster-operator/service/controller/resource/updatemachinedeployments"
	"github.com/giantswarm/cluster-operator/service/controller/resource/workercount"
)

// clusterResourceSetConfig contains necessary dependencies and settings for
// Cluster API's Cluster controller ResourceSet configuration.
type clusterResourceSetConfig struct {
	CertsSearcher certs.Interface
	FileSystem    afero.Fs
	K8sClient     k8sclient.Interface
	Logger        micrologger.Logger
	Tenant        tenantcluster.Interface

	APIIP                      string
	CalicoAddress              string
	CalicoPrefixLength         string
	CertTTL                    string
	ClusterIPRange             string
	DNSIP                      string
	ClusterDomain              string
	NewCommonClusterObjectFunc func() infrastructurev1alpha2.CommonClusterObject
	Provider                   string
	RawAppDefaultConfig        string
	RawAppOverrideConfig       string
	RegistryDomain             string
}

// newClusterResourceSet returns a configured Cluster API's Cluster controller
// ResourceSet.
func newClusterResourceSet(config clusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var haMaster hamaster.Interface
	{
		c := hamaster.Config{
			K8sClient: config.K8sClient,

			Provider: config.Provider,
		}

		haMaster, err = hamaster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appGetter appresource.StateGetter
	{
		c := app.Config{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

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

	var baseDomainResource resource.Interface
	{
		c := basedomain.Config{
			Logger: config.Logger,

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
			HAMaster:  haMaster,
			Logger:    config.Logger,

			APIIP:         config.APIIP,
			CertTTL:       config.CertTTL,
			ClusterDomain: config.ClusterDomain,
			Provider:      config.Provider,
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

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			DNSIP:              config.DNSIP,
			Provider:           config.Provider,
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

			NewCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
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

			NewCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
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

	var keepForInfraRefsResource resource.Interface
	{
		c := keepforinfrarefs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ToObjRef: toClusterObjRef,
		}

		keepForInfraRefsResource, err = keepforinfrarefs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kubeConfigGetter secretresource.StateGetter
	{
		var tenantCluster tenantcluster.Interface
		{
			c := tenantcluster.Config{
				CertsSearcher: config.CertsSearcher,
				Logger:        config.Logger,

				CertID: certs.AppOperatorAPICert,
			}

			tenantCluster, err = tenantcluster.New(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		c := kubeconfig.Config{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient.K8sClient(),
			Logger:        config.Logger,
			Tenant:        tenantCluster,
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

	var releaseVersionsResource resource.Interface
	{
		c := releaseversions.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ToClusterFunc: toClusterFunc,
		}

		releaseVersionsResource, err = releaseversions.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var statusConditionResource resource.Interface
	{
		c := statuscondition.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
			Provider:                   config.Provider,
		}

		statusConditionResource, err = statuscondition.New(c)
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

	var updateG8sControlPlanesResource resource.Interface
	{
		c := updateg8scontrolplanes.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		updateG8sControlPlanesResource, err = updateg8scontrolplanes.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var updateInfraRefsResource resource.Interface
	{
		c := updateinfrarefs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ToObjRef: toClusterObjRef,
			Provider: config.Provider,
		}

		updateInfraRefsResource, err = updateinfrarefs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var updateMachineDeploymentsResource resource.Interface
	{
		c := updatemachinedeployments.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		updateMachineDeploymentsResource, err = updatemachinedeployments.New(c)
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
		releaseVersionsResource,
		tenantClientsResource,
		workerCountResource,

		// Following resources manage resources in the control plane.
		cpNamespaceResource,
		encryptionKeyResource,
		certConfigResource,
		clusterConfigMapResource,
		kubeConfigResource,
		appResource,
		updateG8sControlPlanesResource,
		updateMachineDeploymentsResource,
		updateInfraRefsResource,

		// Following resources manage CR status information.
		clusterIDResource,
		clusterStatusResource,
		statusConditionResource,

		// Following resources manage tenant cluster deletion events.
		cleanupMachineDeployments,
		keepForInfraRefsResource,
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
		return controllercontext.NewContext(ctx, controllercontext.Context{}), nil
	}

	handlesFunc := func(obj interface{}) bool {
		cr, err := key.ToCluster(obj)
		if err != nil {
			return false
		}

		if key.OperatorVersion(&cr) == project.Version() {
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

func toClusterFunc(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return apiv1alpha2.Cluster{}, microerror.Mask(err)
	}

	return cr, nil
}

func toClusterObjRef(obj interface{}) (corev1.ObjectReference, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return corev1.ObjectReference{}, microerror.Mask(err)
	}

	return key.ObjRefFromCluster(cr), nil
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
