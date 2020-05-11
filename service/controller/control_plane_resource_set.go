package controller

import (
	"context"

	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/controller/resource/keepforinfrarefs"
	"github.com/giantswarm/cluster-operator/service/controller/resource/releaseversions"
	"github.com/giantswarm/cluster-operator/service/controller/resource/updateinfrarefs"
)

// controlPlaneResourceSetConfig contains necessary dependencies and settings for
// Cluster API's Cluster controller ResourceSet configuration.
type controlPlaneResourceSetConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	Provider string
}

// newControlPlaneResourceSet returns a configured Control Plane Controller
// ResourceSet.
func newControlPlaneResourceSet(config controlPlaneResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

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
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ToObjRef: toG8sControlPlaneObjRef,
			Provider: config.Provider,
		}

		updateInfraRefsResource, err = updateinfrarefs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var releaseVersionResource resource.Interface
	{
		c := releaseversions.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			ToClusterFunc: newG8sControlPlaneToClusterFunc(config.K8sClient),
		}

		releaseVersionResource, err = releaseversions.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		releaseVersionResource,
		keepForInfraRefsResource,
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

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return controllercontext.NewContext(ctx, controllercontext.Context{}), nil
	}

	handlesFunc := func(obj interface{}) bool {
		cr, err := key.ToG8sControlPlane(obj)
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

func toG8sControlPlaneObjRef(obj interface{}) (corev1.ObjectReference, error) {
	cr, err := key.ToG8sControlPlane(obj)
	if err != nil {
		return corev1.ObjectReference{}, microerror.Mask(err)
	}

	return key.ObjRefFromG8sControlPlane(cr), nil
}
