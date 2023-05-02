package controller

import (
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v8/pkg/controller"
	"github.com/giantswarm/operatorkit/v8/pkg/resource"
	"github.com/giantswarm/operatorkit/v8/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v8/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster/v6/pkg/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v5/pkg/label"
	"github.com/giantswarm/cluster-operator/v5/pkg/project"
	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
	"github.com/giantswarm/cluster-operator/v5/service/controller/resource/controlplanestatus"
	"github.com/giantswarm/cluster-operator/v5/service/controller/resource/deleteinfrarefs"
	"github.com/giantswarm/cluster-operator/v5/service/controller/resource/keepforinfrarefs"
	"github.com/giantswarm/cluster-operator/v5/service/controller/resource/updateinfrarefs"
	"github.com/giantswarm/cluster-operator/v5/service/internal/basedomain"
	"github.com/giantswarm/cluster-operator/v5/service/internal/nodecount"
	"github.com/giantswarm/cluster-operator/v5/service/internal/recorder"
	"github.com/giantswarm/cluster-operator/v5/service/internal/releaseversion"
)

// ControlPlaneConfig contains necessary dependencies and settings for the
// ControlPlane controller implementation.
type ControlPlaneConfig struct {
	BaseDomain     basedomain.Interface
	Event          recorder.Interface
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
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() ctrlClient.Object {
				return new(infrastructurev1alpha3.G8sControlPlane)
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
			Event:     config.Event,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NodeCount: config.NodeCount,
		}

		controlPlaneStatusResource, err = controlplanestatus.New(c)
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
