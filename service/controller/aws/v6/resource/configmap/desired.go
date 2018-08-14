package configmap

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/pkg/v6/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v6/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v6/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := awskey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := awskey.ClusterGuestConfig(customObject)
	configMapValues := configmap.ConfigMapValues{
		ClusterID:    key.ClusterID(clusterGuestConfig),
		Organization: key.ClusterOrganization(clusterGuestConfig),
		// TODO Better comments and issue.
		IngressControllerMigrationEnabled: true,
		IngressControllerServiceEnabled:   false,
		WorkerCount:                       awskey.WorkerCount(customObject),
	}
	desiredConfigMaps, err := r.configMap.GetDesiredState(ctx, configMapValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return desiredConfigMaps, nil
}
