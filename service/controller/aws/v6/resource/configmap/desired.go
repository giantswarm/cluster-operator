package configmap

import (
	"context"
	"fmt"

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
		// Migration is enabled so existing k8scloudconfig resources are
		// replaced.
		IngressControllerMigrationEnabled: true,
		// Proxy protocol is enabled for AWS clusters.
		IngressControllerUseProxyProtocol: true,
		WorkerCount:                       awskey.WorkerCount(customObject),
	}

	r.logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("DEBUG: clusterGuestConfig %#v", clusterGuestConfig))
	r.logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("DEBUG: configMapValues %#v", configMapValues))

	desiredConfigMaps, err := r.configMap.GetDesiredState(ctx, configMapValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", fmt.Sprintf("DEBUG: desired configmaps %#v", desiredConfigMaps))

	return desiredConfigMaps, nil
}
