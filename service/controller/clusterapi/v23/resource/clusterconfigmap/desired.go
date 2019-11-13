package clusterconfigmap

import (
	"context"

	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v23/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var configMap *corev1.ConfigMap
	{
		v := map[string]string{
			"baseDomain":   key.TenantBaseDomain(cr),
			"clusterDNSIP": r.dnsIP,
			"clusterID":    key.ClusterID(&cr),
		}

		b, err := yaml.Marshal(v)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.ClusterConfigMapName(&cr),
				Namespace: key.ClusterID(&cr),
				Labels: map[string]string{
					label.Cluster:      key.ClusterID(&cr),
					label.ManagedBy:    project.Name(),
					label.Organization: key.OrganizationID(&cr),
					label.ServiceType:  label.ServiceTypeManaged,
				},
			},
			Data: map[string]string{
				"values": string(b),
			},
		}
	}

	return []*corev1.ConfigMap{configMap}, nil
}
