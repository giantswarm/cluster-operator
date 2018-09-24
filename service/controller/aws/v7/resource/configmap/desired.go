package configmap

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/pkg/v7/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v7/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v7/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := awskey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := awskey.ClusterGuestConfig(customObject)
	apiDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig := configmap.ClusterConfig{
		APIDomain: apiDomain,
		ClusterID: key.ClusterID(clusterGuestConfig),
	}

	configMapValues := configmap.ConfigMapValues{
		ClusterID:    key.ClusterID(clusterGuestConfig),
		Organization: key.ClusterOrganization(clusterGuestConfig),
		// Migration is enabled so existing k8scloudconfig resources are
		// replaced.
		IngressControllerMigrationEnabled: true,
		// Proxy protocol is enabled for AWS clusters.
		IngressControllerUseProxyProtocol: true,
		WorkerCount:                       awskey.WorkerCount(customObject),
	}
	desiredConfigMaps, err := r.configMap.GetDesiredState(ctx, clusterConfig, configMapValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return desiredConfigMaps, nil
}
