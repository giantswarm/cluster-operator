package configmap

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	pkgkey "github.com/giantswarm/cluster-operator/pkg/v22/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v22/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v22/key"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v22/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v22/key"
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

	// Set provider specific Ingress Controller settings.
	providerValues, err := r.newIngressControllerValues()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMapValues := ConfigMapValues{
		ClusterID: key.ClusterID(&cr),
		CoreDNS: CoreDNSValues{
			CalicoAddress:      r.calicoAddress,
			CalicoPrefixLength: r.calicoPrefixLength,
			ClusterIPRange:     r.clusterIPRange,
			DNSIP:              r.dnsIP,
		},
		IngressController: providerValues,
		Organization:      key.OrganizationID(&cr),
		RegistryDomain:    r.registryDomain,
		WorkerCount:       workerCount(cc.Status.Worker),
	}

	var configMaps []*corev1.ConfigMap

	for _, spec := range newConfigMapSpecs(r.newChartSpecs()) {
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

		configMaps = append(configMaps, newConfigMap(spec))
	}

	return configMaps, nil
}

func (r *Resource) newChartSpecs() []pkgkey.ChartSpec {
	switch r.provider {
	case "aws":
		return append(pkgkey.CommonChartSpecs(), awskey.ChartSpecs()...)
	case "azure":
		return append(pkgkey.CommonChartSpecs(), azurekey.ChartSpecs()...)
	case "kvm":
		return append(pkgkey.CommonChartSpecs(), kvmkey.ChartSpecs()...)
	default:
		return pkgkey.CommonChartSpecs()
	}
}

func (r *Resource) newIngressControllerValues() (IngressControllerValues, error) {
	switch r.provider {
	case "aws":
		return IngressControllerValues{
			// Controller service is disabled because manifest is created by
			// Ignition.
			ControllerServiceEnabled: false,
			// Proxy protocol is enabled for AWS clusters.
			UseProxyProtocol: true,
		}, nil
	case "azure":
		return IngressControllerValues{
			// Controller service is disabled because manifest is not created by
			// Ignition.
			ControllerServiceEnabled: true,
			// Proxy protocol is disabled for Azure clusters.
			UseProxyProtocol: false,
		}, nil
	case "kvm":
		return IngressControllerValues{
			// Controller service is disabled because manifest is created by
			// Ignition.
			ControllerServiceEnabled: false,
			// Proxy protocol is disabled for KVM clusters.
			UseProxyProtocol: false,
		}, nil
	default:
		return IngressControllerValues{}, microerror.Maskf(executionFailedError, "provider %#q not supported")
	}
}

func cidrBlock(address, prefix string) string {
	if address == "" && prefix == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", address, prefix)
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
	calicoCIDRBlock := cidrBlock(configMapValues.CoreDNS.CalicoAddress, configMapValues.CoreDNS.CalicoPrefixLength)

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
					IP: configMapValues.CoreDNS.DNSIP,
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
	values := IngressController{
		Controller: IngressControllerController{
			Replicas: configMapValues.WorkerCount,
			Service: IngressControllerControllerService{
				Enabled: configMapValues.IngressController.ControllerServiceEnabled,
			},
		},
		Global: IngressControllerGlobal{
			Controller: IngressControllerGlobalController{
				UseProxyProtocol: configMapValues.IngressController.UseProxyProtocol,
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

func newConfigMapSpecs(chartSpecs []pkgkey.ChartSpec) []ConfigMapSpec {
	configMapSpecs := make([]ConfigMapSpec, 0)

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

func workerCount(m map[string]controllercontext.ContextStatusWorker) int {
	var n int32

	for _, w := range m {
		n += w.Nodes
	}

	return int(n)
}
