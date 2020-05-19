package releaseversion

import (
	"context"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/cluster-operator/service/internal/releaseversion/internal/cache"
)

type Config struct {
	K8sClient k8sclient.Interface
}

type ReleaseVersion struct {
	k8sClient k8sclient.Interface

	releaseCache *cache.Release
}

func New(c Config) (*ReleaseVersion, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	rv := &ReleaseVersion{
		k8sClient: c.K8sClient,

		releaseCache: cache.NewRelease(),
	}

	return rv, nil
}

func (rv *ReleaseVersion) ReleaseVersioner(ctx context.Context, obj interface{}) (releasev1alpha1.Release, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return releasev1alpha1.Release{}, microerror.Mask(err)
	}

	release, err := rv.cachedRelease(ctx, cr)
	if err != nil {
		return releasev1alpha1.Release{}, microerror.Mask(err)
	}
	return release, nil
}

func (rv *ReleaseVersion) cachedRelease(ctx context.Context, cr metav1.Object) (releasev1alpha1.Release, error) {
	var err error
	var ok bool

	var release releasev1alpha1.Release
	{
		r := rv.releaseCache.Key(ctx, cr)

		if r == "" {
			release, err = rv.lookupReleaseVersions(ctx, cr)
			if err != nil {
				return releasev1alpha1.Release{}, microerror.Mask(err)
			}
		} else {
			release, ok = rv.releaseCache.Get(ctx, r)
			if !ok {
				release, err = rv.lookupReleaseVersions(ctx, cr)
				if err != nil {
					return releasev1alpha1.Release{}, microerror.Mask(err)
				}

				rv.releaseCache.Set(ctx, r, release)
			}
		}
	}

	return release, nil
}

func (rv *ReleaseVersion) lookupReleaseVersions(ctx context.Context, cr metav1.Object) (releasev1alpha1.Release, error) {
	var re releasev1alpha1.Release
	err := rv.k8sClient.CtrlClient().Get(
		ctx,
		types.NamespacedName{Name: key.ReleaseName(key.ReleaseVersion(cr))},
		&re,
	)
	if err != nil {
		return releasev1alpha1.Release{}, microerror.Mask(err)
	}

	return re, nil
}
