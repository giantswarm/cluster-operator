package clusterconfigmap

import (
	"context"
	"strconv"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

type clusterProfile int

const (
	// Here we declare supported cluster profile constants.
	// They are encoded as ordered incrementing numbers so they can be compared
	// with relational operators for equality and ineqality.
	xxs clusterProfile = iota + 1 // worker count == 1
	xs                            // worker count [2-3], max worker CPU cores < 4 or max worker memory < 6 GB
	s                             // worker count > 3, max worker CPU cores < 4 or max worker memory < 6 GB; worker count [2-3], unknown CPU & memory
)

func (r *StateGetter) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Calculating CIDR block for Calico.
	calicoCIDRBlock := key.CIDRBlock(r.calicoAddress, r.calicoPrefixLength)

	// Calculating DNS IP from the IP range so we other operators could use it w/o processing it.
	clusterDNSIP, err := key.DNSIP(r.clusterIPRange)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// controllerServiceEnabled is true for Azure legacy clusters. For AWS and
	// KVM legacy clusters this service is disabled and it is created via
	// ignition.
	var controllerServiceEnabled bool
	{
		if r.provider == "aws" || r.provider == "kvm" || r.provider == "azure" {
			controllerServiceEnabled = false
		} else {
			return nil, microerror.Maskf(executionFailedError, "invalid provider %#q", r.provider)
		}
	}

	// useProxyProtocol is only enabled by default for AWS clusters.
	var useProxyProtocol bool
	{
		if r.provider == "aws" {
			useProxyProtocol = true
		}
	}

	var ingressControllerLegacy bool
	{
		if r.provider == "kvm" {
			ingressControllerLegacy = true
		}
	}

	var determinedTCProfile clusterProfile
	{
		// this is desired, not the current number of tenant cluster worker nodes
		workerCount, err := r.getWorkerCountFunc(obj)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		workerMaxCPUCores, workerMaxCPUCoresKnown, err := r.getWorkerMaxCPUCoresFunc(obj)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		workerMaxMemorySizeGB, workerMaxMemorySizeGBKnown, err := r.getWorkerMaxMemorySizeGBFunc(obj)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if workerCount == 1 {
			determinedTCProfile = xxs
		} else if workerCount < 4 {
			if (workerMaxCPUCoresKnown && workerMaxCPUCores < 4) || (workerMaxMemorySizeGBKnown && workerMaxMemorySizeGB < 6) {
				determinedTCProfile = xs
			} else {
				determinedTCProfile = s
			}
		} else if (workerMaxCPUCoresKnown && workerMaxCPUCores < 4) || (workerMaxMemorySizeGBKnown && workerMaxMemorySizeGB < 6) {
			determinedTCProfile = s
		}
	}

	configMapSpecs := []configMapSpec{
		{
			Name:      key.ClusterConfigMapName(clusterConfig),
			Namespace: key.ClusterID(clusterConfig),
			Values: map[string]interface{}{
				"baseDomain": key.DNSZone(clusterConfig),
				"cluster": map[string]interface{}{
					"calico": map[string]interface{}{
						"CIDR": calicoCIDRBlock,
					},
					"kubernetes": map[string]interface{}{
						"API": map[string]interface{}{
							"clusterIPRange": r.clusterIPRange,
						},
						"DNS": map[string]interface{}{
							"IP": clusterDNSIP,
						},
					},
				},
				"clusterDNSIP": clusterDNSIP,
				"clusterID":    key.ClusterID(clusterConfig),
			},
		},
		{
			Name:      key.IngressControllerConfigMapName,
			Namespace: key.ClusterID(clusterConfig),
			Values: map[string]interface{}{
				"baseDomain": key.DNSZone(clusterConfig),
				"clusterID":  key.ClusterID(clusterConfig),
				"controller": map[string]interface{}{
					"service": map[string]interface{}{
						"enabled": controllerServiceEnabled,
					},
				},
				"configmap": map[string]string{
					"use-proxy-protocol": strconv.FormatBool(useProxyProtocol),
				},
				"ingressController": map[string]interface{}{
					// Legacy flag is set to true so resources created by
					// legacy provider operators are not created.
					"legacy": ingressControllerLegacy,
				},
				"provider": r.provider,
			},
		},
	}

	if determinedTCProfile > 0 {
		for _, configMapSpec := range configMapSpecs {
			_, ok := configMapSpec.Values["cluster"]
			if !ok {
				configMapSpec.Values["cluster"] = make(map[string]interface{})
			}

			clusterMap := configMapSpec.Values["cluster"].(map[string]interface{})
			clusterMap["profile"] = determinedTCProfile
		}
	}

	var configMaps []*corev1.ConfigMap

	for _, spec := range configMapSpecs {
		configMap, err := newConfigMap(clusterConfig, spec)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

func newConfigMap(clusterConfig v1alpha1.ClusterGuestConfig, configMapSpec configMapSpec) (*corev1.ConfigMap, error) {
	yamlValues, err := yaml.Marshal(configMapSpec.Values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapSpec.Name,
			Namespace: configMapSpec.Namespace,
			Labels: map[string]string{
				label.Cluster:      key.ClusterID(clusterConfig),
				label.ManagedBy:    project.Name(),
				label.Organization: clusterConfig.Owner,
				label.ServiceType:  label.ServiceTypeManaged,
			},
		},
		Data: map[string]string{
			"values": string(yamlValues),
		},
	}

	return cm, nil
}
