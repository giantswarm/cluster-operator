package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	var err error
	var chartConfigs []*v1alpha1.ChartConfig

	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorCerts certs.ClusterOperator
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "looking for certificate to connect to the guest cluster")

		operatorCerts, err = r.certsSearcher.SearchClusterOperator(clusterConfig.ClusterID)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found certificate for connecting to the guest cluster")
	}

	var g8sClient versioned.Interface
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating Kubernetes client for the guest cluster")

		var restConfig *rest.Config
		{
			c := k8srestconfig.Config{
				Logger: r.logger,

				Address:   clusterConfig.Domain.API,
				InCluster: false,
				TLS: k8srestconfig.TLSClientConfig{
					CAData:  operatorCerts.APIServer.CA,
					CrtData: operatorCerts.APIServer.Crt,
					KeyData: operatorCerts.APIServer.Key,
				},
			}

			restConfig, err = k8srestconfig.New(c)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		g8sClient, err = versioned.NewForConfig(restConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created Kubernetes client for the guest cluster")
	}

	chartConfigList, err := g8sClient.CoreV1alpha1().ChartConfigs("default").List(apismetav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, item := range chartConfigList.Items {
		chartConfigs = append(chartConfigs, &item)
	}

	return chartConfigs, nil
}
