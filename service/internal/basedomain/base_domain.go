package basedomain

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/internal/basedomain/internal/cache"
)

type Config struct {
	K8sClient k8sclient.Interface
}

type BaseDomain struct {
	k8sClient k8sclient.Interface

	clusterCache *cache.Cluster
}

func New(c Config) (*BaseDomain, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	bd := &BaseDomain{
		k8sClient: c.K8sClient,

		clusterCache: cache.NewCluster(),
	}

	return bd, nil
}

func (bd *BaseDomain) BaseDomain(ctx context.Context, obj interface{}) (string, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	cl, err := bd.cachedCluster(ctx, cr)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return cl.Spec.Cluster.DNS.Domain, nil
}

func (bd *BaseDomain) cachedCluster(ctx context.Context, cr metav1.Object) (infrastructurev1alpha2.AWSCluster, error) {
	var err error
	var ok bool

	var cluster infrastructurev1alpha2.AWSCluster
	{
		ck := bd.clusterCache.Key(ctx, cr)

		if ck == "" {
			cluster, err = bd.lookupCluster(ctx, cr)
			if err != nil {
				return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
			}
		} else {
			cluster, ok = bd.clusterCache.Get(ctx, ck)
			if !ok {
				cluster, err = bd.lookupCluster(ctx, cr)
				if err != nil {
					return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
				}

				bd.clusterCache.Set(ctx, ck, cluster)
			}
		}
	}

	return cluster, nil
}

func (bd *BaseDomain) lookupCluster(ctx context.Context, cr metav1.Object) (infrastructurev1alpha2.AWSCluster, error) {
	var list infrastructurev1alpha2.AWSClusterList

	err := bd.k8sClient.CtrlClient().List(
		ctx,
		&list,
		client.InNamespace(cr.GetNamespace()),
		client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
	)
	if err != nil {
		return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
	}

	if len(list.Items) == 0 {
		return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(notFoundError)
	}
	if len(list.Items) > 1 {
		return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(tooManyCRsError)
	}

	return list.Items[0], nil
}
