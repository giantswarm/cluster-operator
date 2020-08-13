package clusterconfigmap

import (
	"context"
	"fmt"
	"strconv"

	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/v3/pkg/annotation"
	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	bd, err := r.baseDomain.BaseDomain(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var podCIDR string
	{
		podCIDR, err = r.podCIDR.PodCIDR(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// useProxyProtocol is only enabled by default for AWS clusters.
	var useProxyProtocol bool
	{
		if r.provider == "aws" {
			useProxyProtocol = true
		}
	}

	configMapSpecs := []configMapSpec{
		{
			Name:      key.ClusterConfigMapName(&cr),
			Namespace: key.ClusterID(&cr),
			Values: map[string]interface{}{
				"baseDomain": key.TenantEndpoint(&cr, bd),
				"cluster": map[string]interface{}{
					"calico": map[string]interface{}{
						"CIDR": podCIDR,
					},
					"kubernetes": map[string]interface{}{
						"API": map[string]interface{}{
							"clusterIPRange": r.clusterIPRange,
						},
						"DNS": map[string]interface{}{
							"IP": r.dnsIP,
						},
					},
				},
				"clusterDNSIP": r.dnsIP,
				"clusterID":    key.ClusterID(&cr),
			},
		},
		{
			Name:      "ingress-controller-values",
			Namespace: key.ClusterID(&cr),
			Values: map[string]interface{}{
				"baseDomain": key.TenantEndpoint(&cr, bd),
				"clusterID":  key.ClusterID(&cr),
				"configmap": map[string]interface{}{
					"use-proxy-protocol": strconv.FormatBool(useProxyProtocol),
				},
			},
		},
	}

	var configMaps []*corev1.ConfigMap

	for _, spec := range configMapSpecs {
		configMap, err := newConfigMap(cr, spec)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

func newConfigMap(cr apiv1alpha2.Cluster, configMapSpec configMapSpec) (*corev1.ConfigMap, error) {
	yamlValues, err := yaml.Marshal(configMapSpec.Values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapSpec.Name,
			Namespace: configMapSpec.Namespace,
			Annotations: map[string]string{
				annotation.Notes: fmt.Sprintf("DO NOT EDIT. Values managed by %s.", project.Name()),
			},
			Labels: map[string]string{
				label.Cluster:      key.ClusterID(&cr),
				label.ManagedBy:    project.Name(),
				label.Organization: key.OrganizationID(&cr),
				label.ServiceType:  label.ServiceTypeManaged,
			},
		},
		Data: map[string]string{
			"values": string(yamlValues),
		},
	}

	return cm, nil
}
