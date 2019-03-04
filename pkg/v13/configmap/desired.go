package configmap

import (
	"context"
	"encoding/json"
	"math"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v13/key"
)

func (s *Service) GetDesiredState(ctx context.Context, clusterConfig ClusterConfig, configMapValues ConfigMapValues, providerChartSpecs []key.ChartSpec) ([]*corev1.ConfigMap, error) {
	var err error

	desiredConfigMaps := make([]*corev1.ConfigMap, 0)
	configMapSpecs := newConfigMapSpecs(providerChartSpecs)

	for _, spec := range configMapSpecs {
		spec.Labels = newConfigMapLabels(spec, configMapValues, s.projectName)

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
				hasLegacyIC, err := s.hasLegacyIngressController(ctx, spec.ReleaseName, clusterConfig)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				values, err = ingressControllerValues(configMapValues, hasLegacyIC)
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

// hasLegacyIngressController checks if the Ingress Controller deployment
// exists and was created via k8scloudconfig. If so the chart migration
// logic must be enabled.
func (s *Service) hasLegacyIngressController(ctx context.Context, releaseName string, clusterConfig ClusterConfig) (bool, error) {
	tenantK8sClient, err := s.newTenantK8sClient(ctx, clusterConfig)
	if err != nil {
		return false, microerror.Mask(err)
	}

	ingressControllerDeploy, err := tenantK8sClient.Extensions().Deployments(metav1.NamespaceSystem).Get("nginx-ingress-controller", metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// No deployment. So nothing to migrate.
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	// ServiceType label is only present on deployments created via
	// chart-operator.
	serviceType, ok := ingressControllerDeploy.Labels[label.ServiceType]
	if !ok || serviceType != label.ServiceTypeManaged {
		return true, nil
	}

	return false, nil
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

func ingressControllerValues(configMapValues ConfigMapValues, hasLegacyIngressController bool) ([]byte, error) {
	// controllerServiceEnabled needs to be set separately for the chart
	// migration logic but is the reverse of migration enabled.
	controllerServiceEnabled := !configMapValues.IngressController.MigrationEnabled

	migrationEnabled := configMapValues.IngressController.MigrationEnabled
	if migrationEnabled {
		// No legacy ingress controller. So no need for the migration process.
		if hasLegacyIngressController == false {
			migrationEnabled = false
		}
	}

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
				Enabled: controllerServiceEnabled,
			},
		},
		Global: IngressControllerGlobal{
			Controller: IngressControllerGlobalController{
				TempReplicas:     tempReplicas,
				UseProxyProtocol: configMapValues.IngressController.UseProxyProtocol,
			},
			Migration: IngressControllerGlobalMigration{
				Enabled: migrationEnabled,
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
