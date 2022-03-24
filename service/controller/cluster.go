package controller

import (
	"github.com/giantswarm/apiextensions/v6/pkg/annotation"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v7/pkg/controller"
	"github.com/giantswarm/operatorkit/v7/pkg/resource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/crud"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/k8s/configmapresource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/k8s/secretresource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/resource/v4/appresource"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/app"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/appfinalizer"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/appversionlabel"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/certconfig"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/clusterconfigmap"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/clusterid"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/clusterstatus"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/cpnamespace"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/deletecrs"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/deleteinfrarefs"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/keepforcrs"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/keepforinfrarefs"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/kubeconfig"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/statuscondition"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/updateg8scontrolplanes"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/updateinfrarefs"
	"github.com/giantswarm/cluster-operator/v3/service/controller/resource/updatemachinedeployments"
	"github.com/giantswarm/cluster-operator/v3/service/internal/basedomain"
	"github.com/giantswarm/cluster-operator/v3/service/internal/hamaster"
	"github.com/giantswarm/cluster-operator/v3/service/internal/podcidr"
	"github.com/giantswarm/cluster-operator/v3/service/internal/recorder"
	"github.com/giantswarm/cluster-operator/v3/service/internal/releaseversion"
	"github.com/giantswarm/cluster-operator/v3/service/internal/tenantclient"
)

// ClusterConfig contains necessary dependencies and settings for CAPI's Cluster
// CRD controller implementation.
type ClusterConfig struct {
	BaseDomain     basedomain.Interface
	CertsSearcher  certs.Interface
	Event          recorder.Interface
	FileSystem     afero.Fs
	K8sClient      k8sclient.Interface
	Logger         micrologger.Logger
	PodCIDR        podcidr.Interface
	Tenant         tenantcluster.Interface
	ReleaseVersion releaseversion.Interface

	APIIP                      string
	CertTTL                    string
	ClusterIPRange             string
	DNSIP                      string
	ClusterDomain              string
	KiamWatchDogEnabled        bool
	NewCommonClusterObjectFunc func() infrastructurev1alpha3.CommonClusterObject
	Provider                   string
	RawAppDefaultConfig        string
	RawAppOverrideConfig       string
	RegistryDomain             string
}

type Cluster struct {
	*controller.Controller
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	var err error

	var resources []resource.Interface
	{
		resources, err = newClusterResources(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(apiv1alpha3.Cluster)
			},
			Resources: resources,

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/cluster-operator-cluster-controller.
			Name: project.Name() + "-cluster-controller",
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),
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

func newClusterResources(config ClusterConfig) ([]resource.Interface, error) {
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

	var tenantClient tenantclient.Interface
	{
		c := tenantclient.Config{
			BaseDomain:    config.BaseDomain,
			Logger:        config.Logger,
			K8sClient:     config.K8sClient,
			TenantCluster: config.Tenant,
		}

		tenantClient, err = tenantclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appGetter appresource.StateGetter
	{
		c := app.Config{
			G8sClient:      config.K8sClient.G8sClient(),
			K8sClient:      config.K8sClient.K8sClient(),
			Logger:         config.Logger,
			ReleaseVersion: config.ReleaseVersion,

			Provider:             config.Provider,
			KiamWatchDogEnabled:  config.KiamWatchDogEnabled,
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

		c.AllowedAnnotations = []string{
			annotation.LatestConfigMapVersion,
			annotation.LatestSecretVersion,
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

	var appFinalizerResource resource.Interface
	{
		c := appfinalizer.Config{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,
		}

		appFinalizerResource, err = appfinalizer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appVersionLabelResource resource.Interface
	{
		c := appversionlabel.Config{
			G8sClient:      config.K8sClient.G8sClient(),
			Logger:         config.Logger,
			ReleaseVersion: config.ReleaseVersion,
		}

		appVersionLabelResource, err = appversionlabel.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certConfigResource resource.Interface
	{
		c := certconfig.Config{
			BaseDomain:     config.BaseDomain,
			G8sClient:      config.K8sClient.G8sClient(),
			HAMaster:       haMaster,
			Logger:         config.Logger,
			ReleaseVersion: config.ReleaseVersion,

			APIIP:         config.APIIP,
			CertTTL:       config.CertTTL,
			ClusterDomain: config.ClusterDomain,
			Provider:      config.Provider,
		}

		certConfigResource, err = certconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConfigMapGetter configmapresource.StateGetter
	{
		c := clusterconfigmap.Config{
			BaseDomain: config.BaseDomain,
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,
			PodCIDR:    config.PodCIDR,

			ClusterIPRange: config.ClusterIPRange,
			DNSIP:          config.DNSIP,
			Provider:       config.Provider,
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

			AllowedLabels: []string{
				label.AppOperatorWatching,
			},
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
			Provider:                   config.Provider,
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
			Provider:                   config.Provider,
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

	var deleteG8sControlPlaneCRsResource resource.Interface
	{
		c := deletecrs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewObjFunc: func() runtime.Object {
				return &infrastructurev1alpha3.G8sControlPlane{}
			},
			Provider: config.Provider,
		}

		deleteG8sControlPlaneCRsResource, err = deletecrs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deleteMachineDeploymentCRsResource resource.Interface
	{
		c := deletecrs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewObjFunc: func() runtime.Object {
				return &apiv1alpha3.MachineDeployment{}
			},
			Provider: config.Provider,
		}

		deleteMachineDeploymentCRsResource, err = deletecrs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deleteInfraRefsResource resource.Interface
	{
		c := deleteinfrarefs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			Provider: config.Provider,
			ToObjRef: toClusterObjRef,
		}

		deleteInfraRefsResource, err = deleteinfrarefs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keepForG8sControlPlaneCRsResource resource.Interface
	{
		c := keepforcrs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewObjFunc: func() runtime.Object {
				return &infrastructurev1alpha3.G8sControlPlane{}
			},
			Provider: config.Provider,
		}

		keepForG8sControlPlaneCRsResource, err = keepforcrs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keepForMachineDeploymentCRsResource resource.Interface
	{
		c := keepforcrs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			NewObjFunc: func() runtime.Object {
				return &apiv1alpha3.MachineDeployment{}
			},
			Provider: config.Provider,
		}

		keepForMachineDeploymentCRsResource, err = keepforcrs.New(c)
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
			BaseDomain:    config.BaseDomain,
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

	var statusConditionResource resource.Interface
	{
		c := statuscondition.Config{
			Event:          config.Event,
			K8sClient:      config.K8sClient,
			Logger:         config.Logger,
			ReleaseVersion: config.ReleaseVersion,
			TenantClient:   tenantClient,

			NewCommonClusterObjectFunc: config.NewCommonClusterObjectFunc,
			Provider:                   config.Provider,
		}

		statusConditionResource, err = statuscondition.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var updateG8sControlPlanesResource resource.Interface
	{
		c := updateg8scontrolplanes.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			Provider:  config.Provider,
		}

		updateG8sControlPlanesResource, err = updateg8scontrolplanes.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var updateInfraRefsResource resource.Interface
	{
		c := updateinfrarefs.Config{
			K8sClient:      config.K8sClient,
			Logger:         config.Logger,
			ReleaseVersion: config.ReleaseVersion,

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

	resources := []resource.Interface{
		// Following resources manage resources in the control plane.
		cpNamespaceResource,
		certConfigResource,
		clusterConfigMapResource,
		kubeConfigResource,
		appResource,
		appFinalizerResource,
		appVersionLabelResource,
		updateG8sControlPlanesResource,
		updateMachineDeploymentsResource,
		updateInfraRefsResource,

		// Following resources manage CR status information.
		clusterIDResource,
		clusterStatusResource,
		statusConditionResource,

		// Following resources manage tenant cluster deletion events.
		deleteG8sControlPlaneCRsResource,
		deleteMachineDeploymentCRsResource,
		deleteInfraRefsResource,
		keepForG8sControlPlaneCRsResource,
		keepForMachineDeploymentCRsResource,
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

	return resources, nil
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
