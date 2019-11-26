package configmap

import (
	"context"

	"github.com/giantswarm/microerror"

	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/key"
	"github.com/giantswarm/cluster-operator/service/controller/internal/configmap"
	"github.com/giantswarm/cluster-operator/service/controller/key"
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
		ClusterID: key.ClusterID(clusterGuestConfig),
		CoreDNS: configmap.CoreDNSValues{
			CalicoAddress:      r.calicoAddress,
			CalicoPrefixLength: r.calicoPrefixLength,
			ClusterIPRange:     r.clusterIPRange,
		},
		IngressController: configmap.IngressControllerValues{
			// Controller service is disabled because manifest is created by
			// Ignition.
			ControllerServiceEnabled: false,
			// Migration is disabled because AWS is already migrated.
			MigrationEnabled: false,
			// Proxy protocol is enabled for AWS clusters.
			UseProxyProtocol: true,
		},
		Organization:   key.ClusterOrganization(clusterGuestConfig),
		RegistryDomain: r.registryDomain,
		WorkerCount:    awskey.WorkerCount(customObject),
	}
	desiredConfigMaps, err := r.configMap.GetDesiredState(ctx, clusterConfig, configMapValues, awskey.ChartSpecs())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return desiredConfigMaps, nil
}
