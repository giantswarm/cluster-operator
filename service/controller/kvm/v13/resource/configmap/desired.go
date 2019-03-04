package configmap

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/pkg/v12/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v12/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v13/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := kvmkey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := kvmkey.ClusterGuestConfig(customObject)
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
			// Migration is enabled so existing k8scloudconfig resources are
			// replaced.
			MigrationEnabled: true,
			// Proxy protocol is disabled for KVM clusters.
			UseProxyProtocol: false,
		},
		Organization:   key.ClusterOrganization(clusterGuestConfig),
		RegistryDomain: r.registryDomain,
		WorkerCount:    kvmkey.WorkerCount(customObject),
	}
	desiredConfigMaps, err := r.configMap.GetDesiredState(ctx, clusterConfig, configMapValues, kvmkey.ChartSpecs())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return desiredConfigMaps, nil
}
