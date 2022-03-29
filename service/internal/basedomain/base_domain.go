package basedomain

import (
	"context"
	"reflect"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	providerv1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v4/pkg/label"
	"github.com/giantswarm/cluster-operator/v4/service/controller/key"
	"github.com/giantswarm/cluster-operator/v4/service/internal/basedomain/internal/cache"
)

type Config struct {
	K8sClient k8sclient.Interface
	Provider  string
}

type BaseDomain struct {
	k8sClient k8sclient.Interface

	clusterCache *cache.Cluster
	provider     string
}

func New(c Config) (*BaseDomain, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}
	if c.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", c)
	}

	bd := &BaseDomain{
		k8sClient: c.K8sClient,

		clusterCache: cache.NewCluster(),
		provider:     c.Provider,
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

	aws, ok := cl.(infrastructurev1alpha3.AWSCluster)
	if ok {
		return aws.Spec.Cluster.DNS.Domain, nil
	}

	azure, ok := cl.(providerv1alpha1.AzureConfig)
	if ok {
		return azure.Spec.Azure.DNSZones.API.Name, nil
	}

	return "", microerror.Maskf(invalidTypeError, "Cached object was of invalid type %q", reflect.TypeOf(cl))
}

func (bd *BaseDomain) cachedCluster(ctx context.Context, cr metav1.Object) (interface{}, error) {
	var err error
	var ok bool

	var cluster interface{}
	{
		ck := bd.clusterCache.Key(ctx, cr)

		if ck == "" {
			cluster, err = bd.lookupCluster(ctx, cr)
			if err != nil {
				return infrastructurev1alpha3.AWSCluster{}, microerror.Mask(err)
			}
		} else {
			cluster, ok = bd.clusterCache.Get(ctx, ck)
			if !ok {
				cluster, err = bd.lookupCluster(ctx, cr)
				if err != nil {
					return infrastructurev1alpha3.AWSCluster{}, microerror.Mask(err)
				}

				bd.clusterCache.Set(ctx, ck, cluster)
			}
		}
	}

	return cluster, nil
}

func (bd *BaseDomain) lookupCluster(ctx context.Context, cr metav1.Object) (interface{}, error) {
	switch bd.provider {
	case label.ProviderAWS:
		var list infrastructurev1alpha3.AWSClusterList

		err := bd.k8sClient.CtrlClient().List(
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

		err := bd.k8sClient.CtrlClient().List(
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
		return nil, microerror.Maskf(unsupportedProviderError, "Provider %q is unsupported", bd.provider)
	}

	return nil, microerror.Mask(notFoundError)
}
