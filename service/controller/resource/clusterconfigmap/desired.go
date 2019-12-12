package clusterconfigmap

import (
	"context"

	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var ingressControllerReplicas int32
	{
		// We set the number of replicas to the number of worker nodes. This is set
		// by the workercount resource using the current number of nodes from the
		// tenant cluster.
		for _, v := range cc.Status.Worker {
			ingressControllerReplicas += v.Nodes
		}

		// TODO Fallback to desired number of nodes.
	}

	// We limit the number of replicas to 20 as running more than this does
	// not make sense.
	//
	// TODO: Remove Ingress Controller configmap once HPA is enabled by default.
	//
	//	https://github.com/giantswarm/giantswarm/issues/8080
	//
	if ingressControllerReplicas > 20 {
		ingressControllerReplicas = 20
	}

	configMapSpecs := []configMapSpec{
		{
			Name:      key.ClusterConfigMapName(&cr),
			Namespace: key.ClusterID(&cr),
			Values: map[string]interface{}{
				"baseDomain":   key.TenantBaseDomain(cr),
				"clusterDNSIP": r.dnsIP,
				"clusterID":    key.ClusterID(&cr),
			},
		},
		{
			Name:      "ingress-controller-values",
			Namespace: key.ClusterID(&cr),
			Values: map[string]interface{}{
				"baseDomain": key.TenantBaseDomain(cr),
				"clusterID":  key.ClusterID(&cr),
				"ingressController": map[string]interface{}{
					"replicas": ingressControllerReplicas,
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

func newConfigMap(cr cmav1alpha1.Cluster, configMapSpec configMapSpec) (*corev1.ConfigMap, error) {
	yamlValues, err := yaml.Marshal(configMapSpec.Values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapSpec.Name,
			Namespace: configMapSpec.Namespace,
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
