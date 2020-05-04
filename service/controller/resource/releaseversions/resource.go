package releaseversions

import (
	"context"
	"fmt"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

const (
	Name = "releaseversions"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	ToClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error)
}

type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	toClusterFunc func(ctx context.Context, obj interface{}) (apiv1alpha2.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var re releasev1alpha1.Release
	{
		err := r.k8sClient.CtrlClient().Get(
			ctx,
			types.NamespacedName{Name: key.ReleaseName(key.ReleaseVersion(&cr))},
			&re,
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		for _, app := range re.Spec.Apps {
			a := controllercontext.App{
				App:              app.Name,
				ComponentVersion: app.ComponentVersion,
				Version:          app.Version,
			}
			cc.Status.Apps = append(cc.Status.Apps, a)
		}
	}

	{
		if cc.Status.Versions == nil {
			cc.Status.Versions = map[string]string{}
		}
		for _, c := range re.Spec.Components {
			cc.Status.Versions[fmt.Sprintf("%s.giantswarm.io/version", c.Name)] = c.Version
		}

		fmt.Printf("\n")
		fmt.Printf("\n")
		fmt.Printf("\n")
		for k, v := range cc.Status.Versions {
			fmt.Printf("%#v, %#v\n", k, v)
		}
		fmt.Printf("\n")
		fmt.Printf("\n")
		fmt.Printf("\n")
	}

	return nil
}
