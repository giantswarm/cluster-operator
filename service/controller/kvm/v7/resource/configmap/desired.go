package configmap

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/pkg/v7/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v7/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v7/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := kvmkey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := kvmkey.ClusterGuestConfig(customObject)
	guestAPIDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMapConfig := configmap.ConfigMapConfig{
		ClusterID:      key.ClusterID(clusterGuestConfig),
		GuestAPIDomain: guestAPIDomain,
	}

	configMapValues := configmap.ConfigMapValues{
		ClusterID:    key.ClusterID(clusterGuestConfig),
		Organization: key.ClusterOrganization(clusterGuestConfig),
		// Migration is enabled so existing k8scloudconfig resources are
		// replaced.
		IngressControllerMigrationEnabled: true,
		// Proxy protocol is disabled for KVM clusters.
		IngressControllerUseProxyProtocol: false,
		WorkerCount:                       kvmkey.WorkerCount(customObject),
	}
	desiredConfigMaps, err := r.configMap.GetDesiredState(ctx, configMapConfig, configMapValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return desiredConfigMaps, nil
}
