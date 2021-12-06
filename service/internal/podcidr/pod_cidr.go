package podcidr

import (
	"context"
	"reflect"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/internal/podcidr/internal/cache"
)

type Config struct {
	K8sClient k8sclient.Interface

	InstallationCIDR string
	Provider         string
}

type PodCIDR struct {
	k8sClient k8sclient.Interface

	clusterCache *cache.Cluster

	installationCIDR string
	provider         string
}

func New(c Config) (*PodCIDR, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	if c.InstallationCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationCIDR must not be empty", c)
	}
	if c.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", c)
	}

	p := &PodCIDR{
		k8sClient: c.K8sClient,

		clusterCache: cache.NewCluster(),

		installationCIDR: c.InstallationCIDR,
		provider:         c.Provider,
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

	aws, ok := cl.(infrastructurev1alpha3.AWSCluster)
	if ok {
		if aws.Spec.Provider.Pods.CIDRBlock == "" {
			return p.installationCIDR, nil
		} else {
			return aws.Spec.Provider.Pods.CIDRBlock, nil
		}
	}

	azure, ok := cl.(providerv1alpha1.AzureConfig)
	if ok {
		return azure.Spec.Cluster.Calico.Subnet, nil
	}

	return "", microerror.Maskf(invalidTypeError, "Cached object was of invalid type %q", reflect.TypeOf(cl))
}

func (p *PodCIDR) cachedCluster(ctx context.Context, cr metav1.Object) (interface{}, error) {
	var err error
	var ok bool

	var cluster interface{}
	{
		ck := p.clusterCache.Key(ctx, cr)

		if ck == "" {
			cluster, err = p.lookupCluster(ctx, cr)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		} else {
			cluster, ok = p.clusterCache.Get(ctx, ck)
			if !ok {
				cluster, err = p.lookupCluster(ctx, cr)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				p.clusterCache.Set(ctx, ck, cluster)
			}
		}
	}

	return cluster, nil
}

func (p *PodCIDR) lookupCluster(ctx context.Context, cr metav1.Object) (interface{}, error) {
	switch p.provider {
	case label.ProviderAWS:
		var list infrastructurev1alpha3.AWSClusterList

		err := p.k8sClient.CtrlClient().List(
			ctx,
			&list,
			client.InNamespace(cr.GetNamespace()),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if len(list.Items) == 1 {
			return list.Items[0], nil
		}

		if len(list.Items) > 1 {
			return nil, microerror.Mask(tooManyCRsError)
		}
	case label.ProviderAzure:
		var list providerv1alpha1.AzureConfigList

		err := p.k8sClient.CtrlClient().List(
			ctx,
			&list,
			client.InNamespace("default"),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if len(list.Items) == 1 {
			return list.Items[0], nil
		}

		if len(list.Items) > 1 {
			return nil, microerror.Mask(tooManyCRsError)
		}
	default:
		return nil, microerror.Maskf(unsupportedProviderError, "Provider %q is unsupported", p.provider)
	}

	return nil, microerror.Mask(notFoundError)
}
