package configmap

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/pkg/v6/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v6/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v6/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := azurekey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := azurekey.ClusterGuestConfig(customObject)
	configMapValues := configmap.ConfigMapValues{
		ClusterID: key.ClusterID(clusterGuestConfig),
		// Migration is disabled because Azure is already migrated.
		IngressControllerMigrationEnabled: false,
		// Controller Service is enabled because it is created by the chart not
		// by k8scloudconfig.
		IngressControllerServiceEnabled: true,
		Organization:                    key.ClusterOrganization(clusterGuestConfig),
		WorkerCount:                     azurekey.WorkerCount(customObject),
	}
	desiredConfigMaps, err := r.configMap.GetDesiredState(ctx, configMapValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return desiredConfigMaps, nil
}
