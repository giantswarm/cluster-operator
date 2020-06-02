package controller

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster/v2/pkg/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/controller/resource/controlplanestatus"
	"github.com/giantswarm/cluster-operator/service/controller/resource/deleteinfrarefs"
	"github.com/giantswarm/cluster-operator/service/controller/resource/keepforinfrarefs"
	"github.com/giantswarm/cluster-operator/service/controller/resource/tenantclients"
	"github.com/giantswarm/cluster-operator/service/controller/resource/updateinfrarefs"
	"github.com/giantswarm/cluster-operator/service/internal/basedomain"
	"github.com/giantswarm/cluster-operator/service/internal/nodecount"
	"github.com/giantswarm/cluster-operator/service/internal/releaseversion"
)

// ControlPlaneConfig contains necessary dependencies and settings for the
// ControlPlane controller implementation.
type ControlPlaneConfig struct {
	BaseDomain     basedomain.Interface
	K8sClient      k8sclient.Interface
	Logger         micrologger.Logger
	NodeCount      nodecount.Interface
	Tenant         tenantcluster.Interface
	ReleaseVersion releaseversion.Interface

	Provider string
}

type ControlPlane struct {
	*controller.Controller
}

func NewControlPlane(config ControlPlaneConfig) (*ControlPlane, error) {
	var err error

	var resources []resource.Interface
	{
		resources, err = newControlPlaneResources(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var controlPlaneController *controller.Controller
	{
		c := controller.Config{
			InitCtx: func(ctx context.Context, obj interface{}) (context.Context, error) {
				return controllercontext.NewContext(ctx, controllercontext.Context{}), nil
			},
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(infrastructurev1alpha2.G8sControlPlane)
			},
			Resources: resources,

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/cluster-operator-control-plane-controller.
			Name: project.Name() + "-control-plane-controller",
			Selector: labels.SelectorFromSet(map[string]string{
				label.OperatorVersion: project.Version(),
			}),
		}

		controlPlaneController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &ControlPlane{
		Controller: controlPlaneController,
	}

	return c, nil
}

func newControlPlaneResources(config ControlPlaneConfig) ([]resource.Interface, error) {
	var err error

	var controlPlaneStatusResource resource.Interface
	{
		c := controlplanestatus.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NodeCount: config.NodeCount,
		}

		controlPlaneStatusResource, err = controlplanestatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClientsResource resource.Interface
	{
		c := tenantclients.Config{
			BaseDomain:    config.BaseDomain,
			Logger:        config.Logger,
			Tenant:        config.Tenant,
			ToClusterFunc: newG8sControlPlaneToClusterFunc(config.K8sClient),
		}

		tenantClientsResource, err = tenantclients.New(c)
  
	var deleteInfraRefsResource resource.Interface
	{
		c := deleteinfrarefs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ToObjRef: toG8sControlPlaneObjRef,
		}

		deleteInfraRefsResource, err = deleteinfrarefs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var keepForInfraRefsResource resource.Interface
	{
		c := keepforinfrarefs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ToObjRef: toG8sControlPlaneObjRef,
		}

		keepForInfraRefsResource, err = keepforinfrarefs.New(c)
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

			ToObjRef: toG8sControlPlaneObjRef,
			Provider: config.Provider,
		}

		updateInfraRefsResource, err = updateinfrarefs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		tenantClientsResource,
		// Following resources manage CR status information. Note that
		// keepForInfraRefsResource needs to run before
		// controlPlaneStatusResource because keepForInfraRefsResource keeps
		// finalizers where controlPlaneStatusResource does not.
		deleteInfraRefsResource,
		keepForInfraRefsResource,
		controlPlaneStatusResource,

		// Following resources manage resources in the control plane.
		updateInfraRefsResource,
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

func toG8sControlPlaneObjRef(obj interface{}) (corev1.ObjectReference, error) {
	cr, err := key.ToG8sControlPlane(obj)
	if err != nil {
		return corev1.ObjectReference{}, microerror.Mask(err)
	}

	return key.ObjRefFromG8sControlPlane(cr), nil
}

func newG8sControlPlaneToClusterFunc(k8sClient k8sclient.Interface) func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error) {
	return func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error) {
		cr := &apiv1alpha2.Cluster{}
		{
			cp, err := key.ToG8sControlPlane(obj)
			if err != nil {
				return apiv1alpha2.Cluster{}, microerror.Mask(err)
			}

			// Note that we cannot use a key function here because we do not need to
			// fetch the Control Plane again. We need to lookup the Cluster CR based
			// on the G8sControlPlane CR. This is why we use types.NamespacedName here
			// explicitly.
			err = k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: key.ClusterID(&cp), Namespace: cp.Namespace}, cr)
			if err != nil {
				return apiv1alpha2.Cluster{}, microerror.Mask(err)
			}
		}

		return *cr, nil
	}
}
