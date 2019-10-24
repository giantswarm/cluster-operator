package clusterconfigmap

import (
	"context"

	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/pkg/v21/key"
)

func (r *StateGetter) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMapName := key.ClusterConfigMapName(clusterConfig)

	// Calculating DNS IP from the IP range so we other operators could use it w/o processing it.
	clusterDNSIP, err := key.DNSIP(r.clusterIPRange)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	values := map[string]string{
		"baseDomain":   key.DNSZone(clusterConfig),
		"clusterDNSIP": clusterDNSIP,
		"clusterID":    key.ClusterID(clusterConfig),
	}

	yamlValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: clusterConfig.ID,
			Labels: map[string]string{
				label.Cluster:      clusterConfig.ID,
				label.ManagedBy:    project.Name(),
				label.Organization: clusterConfig.Owner,
				label.ServiceType:  label.ServiceTypeManaged,
			},
		},
		Data: map[string]string{
			"values": string(yamlValues),
		},
	}

	return []*corev1.ConfigMap{&cm}, nil
}
