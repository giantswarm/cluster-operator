package configmap

import (
	"context"
	"encoding/json"
	"math"

	"github.com/giantswarm/azure-operator/service/controller/v6/controllercontext"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/pkg/v10/configmap"
	"github.com/giantswarm/cluster-operator/pkg/v19/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
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

	desiredConfigMaps := make([]*corev1.ConfigMap, 0)
	configMapSpecs := newConfigMapSpecs(providerChartSpecs)

	for _, spec := range configMapSpecs {
		spec.Labels = newConfigMapLabels(spec, configMapValues, project.Name())

		// Values are only set for app configmaps.
		if spec.Type == label.ConfigMapTypeApp {
			values := []byte{}

			switch spec.App {
			case "cert-exporter":
				values, err = exporterValues(configMapValues)
				if err != nil {
					return nil, microerror.Mask(err)
				}
			case "cluster-autoscaler":
				values, err = clusterAutoscalerValues(configMapValues)
				if err != nil {
					return nil, microerror.Mask(err)
				}
			case "coredns":
				values, err = coreDNSValues(configMapValues)
				if err != nil {
					return nil, microerror.Mask(err)
				}
			case "net-exporter":
				values, err = exporterValues(configMapValues)
				if err != nil {
					return nil, microerror.Mask(err)
				}
			case "nginx-ingress-controller":
				values, err = ingressControllerValues(configMapValues)
				if err != nil {
					return nil, microerror.Mask(err)
				}
			default:
				values, err = defaultValues(configMapValues)
				if err != nil {
					return nil, microerror.Mask(err)
				}
			}

			spec.ValuesJSON = string(values)
		}

		desiredConfigMaps = append(desiredConfigMaps, newConfigMap(spec))
	}

	return desiredConfigMaps, nil
}

func clusterAutoscalerValues(configMapValues ConfigMapValues) ([]byte, error) {
	values := ClusterAutoscaler{
		Cluster: ClusterAutoscalerCluster{
			ID: configMapValues.ClusterID,
		},
		Image: Image{
			Registry: configMapValues.RegistryDomain,
		},
	}
	json, err := json.Marshal(values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return json, nil
}

func coreDNSValues(configMapValues ConfigMapValues) ([]byte, error) {
	calicoCIDRBlock := key.CIDRBlock(configMapValues.CoreDNS.CalicoAddress, configMapValues.CoreDNS.CalicoPrefixLength)
	DNSIP, err := key.DNSIP(configMapValues.CoreDNS.ClusterIPRange)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	values := CoreDNS{
		Cluster: CoreDNSCluster{
			Calico: CoreDNSClusterCalico{
				CIDR: calicoCIDRBlock,
			},
			Kubernetes: CoreDNSClusterKubernetes{
				API: CoreDNSClusterKubernetesAPI{
					ClusterIPRange: configMapValues.CoreDNS.ClusterIPRange,
				},
				DNS: CoreDNSClusterKubernetesDNS{
					IP: DNSIP,
				},
			},
		},
		Image: Image{
			Registry: configMapValues.RegistryDomain,
		},
	}
	json, err := json.Marshal(values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return json, nil
}

func defaultValues(configMapValues ConfigMapValues) ([]byte, error) {
	values := DefaultConfigMap{
		Image: Image{
			Registry: configMapValues.RegistryDomain,
		},
	}
	json, err := json.Marshal(values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return json, nil
}

func exporterValues(configMapValues ConfigMapValues) ([]byte, error) {
	values := ExporterValues{
		Namespace: metav1.NamespaceSystem,
	}
	json, err := json.Marshal(values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return json, nil
}

func ingressControllerValues(configMapValues ConfigMapValues) ([]byte, error) {
	// tempReplicas is set to 50% of the worker count to ensure all pods can be
	// scheduled.
	tempReplicas, err := setIngressControllerTempReplicas(configMapValues.WorkerCount)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	values := IngressController{
		Controller: IngressControllerController{
			Replicas: configMapValues.WorkerCount,
			Service: IngressControllerControllerService{
				Enabled: configMapValues.IngressController.ControllerServiceEnabled,
			},
		},
		Global: IngressControllerGlobal{
			Controller: IngressControllerGlobalController{
				TempReplicas:     tempReplicas,
				UseProxyProtocol: configMapValues.IngressController.UseProxyProtocol,
			},
			Migration: IngressControllerGlobalMigration{
				Enabled: configMapValues.IngressController.MigrationEnabled,
			},
		},
		Image: Image{
			Registry: configMapValues.RegistryDomain,
		},
	}
	json, err := json.Marshal(values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return json, nil
}

func newConfigMap(configMapSpec ConfigMapSpec) *corev1.ConfigMap {
	data := make(map[string]string)

	// Values are only set for app configmaps.
	if configMapSpec.ValuesJSON != "" {
		data["values.json"] = configMapSpec.ValuesJSON
	}

	newConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapSpec.Name,
			Namespace: configMapSpec.Namespace,
			Labels:    configMapSpec.Labels,
		},
		Data: data,
	}

	return newConfigMap
}

func newConfigMapLabels(configMapSpec ConfigMapSpec, configMapValues ConfigMapValues, projectName string) map[string]string {
	return map[string]string{
		label.App:           configMapSpec.App,
		label.Cluster:       configMapValues.ClusterID,
		label.ConfigMapType: configMapSpec.Type,
		label.ManagedBy:     projectName,
		label.Organization:  configMapValues.Organization,
		label.ServiceType:   label.ServiceTypeManaged,
	}
}

func newConfigMapSpecs(providerChartSpecs []key.ChartSpec) []ConfigMapSpec {
	configMapSpecs := make([]ConfigMapSpec, 0)

	// Add common and provider specific chart specs.
	chartSpecs := key.CommonChartSpecs()
	chartSpecs = append(chartSpecs, providerChartSpecs...)

	for _, chartSpec := range chartSpecs {
		if chartSpec.ConfigMapName != "" {
			configMapSpec := ConfigMapSpec{
				App:         chartSpec.AppName,
				Name:        chartSpec.ConfigMapName,
				Namespace:   chartSpec.Namespace,
				ReleaseName: chartSpec.ReleaseName,
				Type:        label.ConfigMapTypeApp,
			}

			configMapSpecs = append(configMapSpecs, configMapSpec)
		}

		if chartSpec.UserConfigMapName != "" {
			configMapSpec := ConfigMapSpec{
				App:         chartSpec.AppName,
				Name:        chartSpec.UserConfigMapName,
				Namespace:   chartSpec.Namespace,
				ReleaseName: chartSpec.ReleaseName,
				Type:        label.ConfigMapTypeUser,
			}

			configMapSpecs = append(configMapSpecs, configMapSpec)
		}
	}

	return configMapSpecs
}

// setIngressControllerTempReplicas sets the temp replicas to 50% of the worker
// count to ensure all pods can be scheduled.
func setIngressControllerTempReplicas(workerCount int) (int, error) {
	if workerCount == 0 {
		return 0, microerror.Maskf(invalidExecutionError, "worker count must not be 0")
	}

	tempReplicas := float64(workerCount) * float64(0.5)

	return int(math.Round(tempReplicas)), nil
}
