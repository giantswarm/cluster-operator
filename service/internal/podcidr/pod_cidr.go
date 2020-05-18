package podcidr

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/internal/podcidr/internal/cache"
)

type Config struct {
	K8sClient k8sclient.Interface

	InstallationCIDR string
}

type PodCIDR struct {
	k8sClient k8sclient.Interface

	clusterCache *cache.Cluster

	installationCIDR string
}

func New(c Config) (*PodCIDR, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	if c.InstallationCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationCIDR must not be empty", c)
	}

	p := &PodCIDR{
		k8sClient: c.K8sClient,

		clusterCache: cache.NewCluster(),

		installationCIDR: c.InstallationCIDR,
	}

	return p, nil
}

func (p *PodCIDR) PodCIDR(ctx context.Context, obj interface{}) (string, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	cl, err := p.cachedCluster(ctx, cr)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var podCIDR string
	if cl.Spec.Provider.Pods.CIDRBlock == "" {
		podCIDR = p.installationCIDR
	} else {
		podCIDR = cl.Spec.Provider.Pods.CIDRBlock
	}

	return podCIDR, nil
}

func (p *PodCIDR) cachedCluster(ctx context.Context, cr metav1.Object) (infrastructurev1alpha2.AWSCluster, error) {
	var err error
	var ok bool

	var cluster infrastructurev1alpha2.AWSCluster
	{
		ck := p.clusterCache.Key(ctx, cr)

		if ck == "" {
			cluster, err = p.lookupCluster(ctx, cr)
			if err != nil {
				return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
			}
		} else {
			cluster, ok = p.clusterCache.Get(ctx, ck)
			if !ok {
				cluster, err = p.lookupCluster(ctx, cr)
				if err != nil {
					return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
				}

				p.clusterCache.Set(ctx, ck, cluster)
			}
		}
	}

	return cluster, nil
}

func (p *PodCIDR) lookupCluster(ctx context.Context, cr metav1.Object) (infrastructurev1alpha2.AWSCluster, error) {
	var list infrastructurev1alpha2.AWSClusterList

	err := p.k8sClient.CtrlClient().List(
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
