package configmap

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v7/key"
)

const (
	coreDNSJSON = `
  {
    "cluster": {
      "calico": {
        "cidr": "172.20.0.0/16"
      },
      "kubernetes": {
        "api": {
          "clusterIPRange": "172.31.0.0/16"
        },
        "dns": {
          "ip": "172.31.0.10"
        }
      }
    },
    "image": {
      "registry": "quay.io"
    }
  }
`

	basicMatchJSON = `
	{
		"controller": {
			"replicas": 3,
			"service": {
				"enabled": false
			}
		},
		"global": {
			"controller": {
				"tempReplicas": 2,
				"useProxyProtocol": true
			},
			"migration": {
				"enabled": true
			}
		},
		"image": {
			"registry": "quay.io"
		}
	}
	`
	differentWorkerCountJSON = `
	{
		"controller": {
			"replicas": 7,
			"service": {
				"enabled": false
			}
		},
		"global": {
			"controller": {
				"tempReplicas": 4,
				"useProxyProtocol": true
			},
			"migration": {
				"enabled": true
			}
		},
		"image": {
			"registry": "quay.io"
		}
	}
	`
	differentSettingsJSON = `
	{
		"controller": {
			"replicas": 1,
			"service": {
				"enabled": true
			}
		},
		"global": {
			"controller": {
				"tempReplicas": 1,
				"useProxyProtocol": false
			},
			"migration": {
				"enabled": false
			}
		},
		"image": {
			"registry": "quay.io"
		}
	}
	`
	alreadyMigratedJSON = `
	{
		"controller": {
			"replicas": 3,
			"service": {
				"enabled": false
			}
		},
		"global": {
			"controller": {
				"tempReplicas": 2,
				"useProxyProtocol": false
			},
			"migration": {
				"enabled": false
			}
		},
		"image": {
			"registry": "quay.io"
		}
	}
	`
)

func Test_ConfigMap_GetDesiredState(t *testing.T) {
	testCases := []struct {
		name                            string
		configMapConfig                 ClusterConfig
		configMapValues                 ConfigMapValues
		ingressControllerReleasePresent bool
		expectedConfigMaps              []*corev1.ConfigMap
	}{
		{
			name: "case 0: basic match",
			configMapConfig: ClusterConfig{
				APIDomain:  "5xchu.aws.giantswarm.io",
				ClusterID:  "5xchu",
				Namespaces: []string{},
			},
			configMapValues: ConfigMapValues{
				ClusterID:                         "5xchu",
				IngressControllerMigrationEnabled: true,
				IngressControllerUseProxyProtocol: true,
				Organization:                      "giantswarm",
				WorkerCount:                       3,
			},
			ingressControllerReleasePresent: false,
			expectedConfigMaps: []*corev1.ConfigMap{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cert-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "cert-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"namespace\":\"kube-system\"}",
					},
				},
				/*
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "coredns-values",
							Namespace: metav1.NamespaceSystem,
							Labels: map[string]string{
								label.App:          "coredns",
								label.Cluster:      "5xchu",
								label.ManagedBy:    "cluster-operator",
								label.Organization: "giantswarm",
								label.ServiceType:  "managed",
							},
						},
						Data: map[string]string{
							"values.json": coreDNSJSON,
						},
					},
				*/
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-ingress-controller-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "nginx-ingress-controller",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": basicMatchJSON,
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-state-metrics-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "kube-state-metrics",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"image\":{\"registry\":\"quay.io\"}}",
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "net-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "net-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"namespace\":\"kube-system\"}",
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "node-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "node-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"image\":{\"registry\":\"quay.io\"}}",
					},
				},
			},
		},
		{
			name: "case 1: different worker count",
			configMapConfig: ClusterConfig{
				APIDomain:  "5xchu.aws.giantswarm.io",
				ClusterID:  "5xchu",
				Namespaces: []string{},
			},
			configMapValues: ConfigMapValues{
				ClusterID:                         "5xchu",
				Organization:                      "giantswarm",
				IngressControllerMigrationEnabled: true,
				IngressControllerUseProxyProtocol: true,
				WorkerCount:                       7,
			},
			ingressControllerReleasePresent: false,
			expectedConfigMaps: []*corev1.ConfigMap{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cert-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "cert-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"namespace\":\"kube-system\"}",
					},
				},
				/*
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "coredns-values",
							Namespace: metav1.NamespaceSystem,
							Labels: map[string]string{
								label.App:          "coredns",
								label.Cluster:      "5xchu",
								label.ManagedBy:    "cluster-operator",
								label.Organization: "giantswarm",
								label.ServiceType:  "managed",
							},
						},
						Data: map[string]string{
							"values.json": coreDNSJSON,
						},
					},
				*/
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-ingress-controller-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "nginx-ingress-controller",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": differentWorkerCountJSON,
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-state-metrics-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "kube-state-metrics",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"image\":{\"registry\":\"quay.io\"}}",
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "net-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "net-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"namespace\":\"kube-system\"}",
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "node-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "node-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"image\":{\"registry\":\"quay.io\"}}",
					},
				},
			},
		},
		{
			name: "case 2: different ingress controller settings",
			configMapConfig: ClusterConfig{
				APIDomain:  "5xchu.aws.giantswarm.io",
				ClusterID:  "5xchu",
				Namespaces: []string{},
			},
			configMapValues: ConfigMapValues{
				ClusterID:                         "5xchu",
				IngressControllerMigrationEnabled: false,
				IngressControllerUseProxyProtocol: false,
				Organization:                      "giantswarm",
				WorkerCount:                       1,
			},
			ingressControllerReleasePresent: false,
			expectedConfigMaps: []*corev1.ConfigMap{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cert-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "cert-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"namespace\":\"kube-system\"}",
					},
				},
				/*
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "coredns-values",
							Namespace: metav1.NamespaceSystem,
							Labels: map[string]string{
								label.App:          "coredns",
								label.Cluster:      "5xchu",
								label.ManagedBy:    "cluster-operator",
								label.Organization: "giantswarm",
								label.ServiceType:  "managed",
							},
						},
						Data: map[string]string{
							"values.json": coreDNSJSON,
						},
					},
				*/
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-ingress-controller-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "nginx-ingress-controller",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": differentSettingsJSON,
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-state-metrics-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "kube-state-metrics",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"image\":{\"registry\":\"quay.io\"}}",
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "net-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "net-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"namespace\":\"kube-system\"}",
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "node-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "node-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"image\":{\"registry\":\"quay.io\"}}",
					},
				},
			},
		},
		{
			name: "case 3: ingress controller already migrated",
			configMapConfig: ClusterConfig{
				APIDomain:  "5xchu.aws.giantswarm.io",
				ClusterID:  "5xchu",
				Namespaces: []string{},
			},
			configMapValues: ConfigMapValues{
				ClusterID:                         "5xchu",
				IngressControllerMigrationEnabled: true,
				Organization:                      "giantswarm",
				WorkerCount:                       3,
			},
			ingressControllerReleasePresent: true,
			expectedConfigMaps: []*corev1.ConfigMap{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cert-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "cert-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"namespace\":\"kube-system\"}",
					},
				},
				/*
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "coredns-values",
							Namespace: metav1.NamespaceSystem,
							Labels: map[string]string{
								label.App:          "coredns",
								label.Cluster:      "5xchu",
								label.ManagedBy:    "cluster-operator",
								label.Organization: "giantswarm",
								label.ServiceType:  "managed",
							},
						},
						Data: map[string]string{
							"values.json": coreDNSJSON,
						},
					},
				*/
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-ingress-controller-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "nginx-ingress-controller",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": alreadyMigratedJSON,
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-state-metrics-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "kube-state-metrics",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"image\":{\"registry\":\"quay.io\"}}",
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "net-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "net-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"namespace\":\"kube-system\"}",
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "node-exporter-values",
						Namespace: metav1.NamespaceSystem,
						Labels: map[string]string{
							label.App:          "node-exporter",
							label.Cluster:      "5xchu",
							label.ManagedBy:    "cluster-operator",
							label.Organization: "giantswarm",
							label.ServiceType:  "managed",
						},
					},
					Data: map[string]string{
						"values.json": "{\"image\":{\"registry\":\"quay.io\"}}",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helmClient := &helmMock{}
			if !tc.ingressControllerReleasePresent {
				helmClient.defaultError = microerror.Newf("No such release: nginx-ingress-controller")
			}

			c := Config{
				Logger: microloggertest.New(),
				Tenant: &tenantMock{
					fakeTenantHelmClient: helmClient,
				},

				CalicoAddress:      "172.20.0.0",
				CalicoPrefixLength: "16",
				ClusterIPRange:     "172.31.0.0/16",
				ProjectName:        "cluster-operator",
				RegistryDomain:     "quay.io",
			}
			newService, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			configMaps, err := newService.GetDesiredState(context.TODO(), tc.configMapConfig, tc.configMapValues, []key.ChartSpec{})
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if len(configMaps) != len(tc.expectedConfigMaps) {
				t.Fatal("expected", len(tc.expectedConfigMaps), "got", len(configMaps))
			}

			for _, expectedConfigMap := range tc.expectedConfigMaps {
				configMap, err := getConfigMapByNameAndNamespace(configMaps, expectedConfigMap.Name, expectedConfigMap.Namespace)
				if IsNotFound(err) {
					t.Fatalf("expected config map '%s' not found", expectedConfigMap.Name)
				} else if err != nil {
					t.Fatalf("expected nil, got %#v", err)
				}

				if !reflect.DeepEqual(configMap.ObjectMeta.Labels, expectedConfigMap.ObjectMeta.Labels) {
					t.Fatalf("expected config map labels %#v, got %#v", expectedConfigMap.ObjectMeta.Labels, configMap.ObjectMeta.Labels)
				}

				for expectedKey, expectedValues := range expectedConfigMap.Data {
					values, ok := configMap.Data[expectedKey]
					if !ok {
						t.Fatalf("expected key '%s' not found", expectedKey)
					}

					equalValues, err := compareJSON(expectedValues, values)
					if err != nil {
						t.Fatal("expected", nil, "got", err)
					}
					if !equalValues {
						t.Fatal("expected", expectedValues, "got", values)
					}
				}
			}
		})
	}
}

func Test_ConfigMap_newConfigMap(t *testing.T) {
	testCases := []struct {
		name              string
		configMapSpec     ConfigMapSpec
		expectedConfigMap *corev1.ConfigMap
	}{
		{
			name: "case 0: basic match with no labels or values",
			configMapSpec: ConfigMapSpec{
				App:       "test-app",
				Name:      "test-app-values",
				Namespace: metav1.NamespaceSystem,
			},
			expectedConfigMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app-values",
					Namespace: metav1.NamespaceSystem,
				},
				Data: map[string]string{},
			},
		},
		{
			name: "case 1: has labels but no values",
			configMapSpec: ConfigMapSpec{
				App:       "test-app",
				Name:      "test-app-values",
				Namespace: metav1.NamespaceSystem,
				Labels: map[string]string{
					"app": "test-app",
					"giantswarm.io/cluster": "5xchu",
				},
			},
			expectedConfigMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app-values",
					Namespace: metav1.NamespaceSystem,
					Labels: map[string]string{
						"app": "test-app",
						"giantswarm.io/cluster": "5xchu",
					},
				},
				Data: map[string]string{},
			},
		},
		{
			name: "case 2: has labels and values",
			configMapSpec: ConfigMapSpec{
				App:       "test-app",
				Name:      "test-app-values",
				Namespace: metav1.NamespaceSystem,
				Labels: map[string]string{
					"app": "test-app",
					"giantswarm.io/cluster": "5xchu",
				},
				ValuesJSON: "{\"image\":{\"registry\":\"quay.io\"}}",
			},
			expectedConfigMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app-values",
					Namespace: metav1.NamespaceSystem,
					Labels: map[string]string{
						"app": "test-app",
						"giantswarm.io/cluster": "5xchu",
					},
				},
				Data: map[string]string{
					"values.json": "{\"image\":{\"registry\":\"quay.io\"}}",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configMap := newConfigMap(tc.configMapSpec)

			if configMap.Name != tc.expectedConfigMap.Name {
				t.Fatalf("expected name %#q got %#q", tc.expectedConfigMap.Name, configMap.Name)
			}
			if configMap.Namespace != tc.expectedConfigMap.Namespace {
				t.Fatalf("expected namespace %#q got %#q", tc.expectedConfigMap.Namespace, configMap.Namespace)
			}
			if !reflect.DeepEqual(configMap.Labels, tc.expectedConfigMap.Labels) {
				t.Fatalf("expected labels %#v got %#v", tc.expectedConfigMap.Labels, configMap.Labels)
			}

			if !reflect.DeepEqual(configMap.Data, tc.expectedConfigMap.Data) {
				t.Fatalf("expected data %#v got %#v", tc.expectedConfigMap.Data, configMap.Data)
			}
		})
	}
}

func Test_ConfigMap_coreDNSValues(t *testing.T) {
	testCases := []struct {
		name               string
		configMapValues    ConfigMapValues
		errorMatcher       func(error) bool
		expectedValuesJSON string
	}{
		{
			name: "case 0: basic match",
			configMapValues: ConfigMapValues{
				CalicoAddress:      "172.20.0.0",
				CalicoPrefixLength: "16",
				ClusterIPRange:     "172.31.0.0/16",
				RegistryDomain:     "quay.io",
			},
			expectedValuesJSON: coreDNSJSON,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			values, err := coreDNSValues(tc.configMapValues)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			equalValues, err := compareValuesJSON(tc.expectedValuesJSON, values)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			if !equalValues {
				t.Fatal("expected", tc.expectedValuesJSON, "got", string(values))
			}
		})
	}
}

func Test_ConfigMap_defaultValues(t *testing.T) {
	testCases := []struct {
		name               string
		configMapValues    ConfigMapValues
		errorMatcher       func(error) bool
		expectedValuesJSON string
	}{
		{
			name: "case 0: basic match",
			configMapValues: ConfigMapValues{
				RegistryDomain: "quay.io",
			},
			expectedValuesJSON: `{ "image": { "registry": "quay.io" } }`,
		},
		{
			name: "case 1: different registry",
			configMapValues: ConfigMapValues{
				RegistryDomain: "gcr.io",
			},
			expectedValuesJSON: `{ "image": { "registry": "gcr.io" } }`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			values, err := defaultValues(tc.configMapValues)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			equalValues, err := compareValuesJSON(tc.expectedValuesJSON, values)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			if !equalValues {
				t.Fatal("expected", tc.expectedValuesJSON, "got", string(values))
			}
		})
	}
}

func Test_ConfigMap_exporterValues(t *testing.T) {
	testCases := []struct {
		name               string
		configMapValues    ConfigMapValues
		errorMatcher       func(error) bool
		expectedValuesJSON string
	}{
		{
			name:               "case 0: basic match",
			configMapValues:    ConfigMapValues{},
			expectedValuesJSON: `{ "namespace": "kube-system" }`,
		},
		{
			name:               "case 1: different registry",
			configMapValues:    ConfigMapValues{},
			expectedValuesJSON: `{ "namespace": "kube-system" }`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			values, err := exporterValues(tc.configMapValues)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			equalValues, err := compareValuesJSON(tc.expectedValuesJSON, values)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			if !equalValues {
				t.Fatalf("expected JSON: \n %s \n got JSON: \n %s", tc.expectedValuesJSON, values)
			}
		})
	}
}

func Test_ConfigMap_ingressControllerValues(t *testing.T) {
	testCases := []struct {
		name               string
		configMapValues    ConfigMapValues
		releaseExists      bool
		errorMatcher       func(error) bool
		expectedValuesJSON string
	}{
		{
			name: "case 0: basic match",
			configMapValues: ConfigMapValues{
				IngressControllerMigrationEnabled: true,
				IngressControllerUseProxyProtocol: true,
				RegistryDomain:                    "quay.io",
				WorkerCount:                       3,
			},
			releaseExists:      false,
			expectedValuesJSON: basicMatchJSON,
		},
		{
			name: "case 1: different worker count",
			configMapValues: ConfigMapValues{
				IngressControllerMigrationEnabled: true,
				IngressControllerUseProxyProtocol: true,
				RegistryDomain:                    "quay.io",
				WorkerCount:                       7,
			},
			releaseExists:      false,
			expectedValuesJSON: differentWorkerCountJSON,
		},
		{
			name: "case 2: different settings",
			configMapValues: ConfigMapValues{
				IngressControllerMigrationEnabled: false,
				IngressControllerUseProxyProtocol: false,
				RegistryDomain:                    "quay.io",
				WorkerCount:                       1,
			},
			releaseExists:      false,
			expectedValuesJSON: differentSettingsJSON,
		},
		{
			name: "case 3: already migrated",
			configMapValues: ConfigMapValues{
				IngressControllerMigrationEnabled: true,
				IngressControllerUseProxyProtocol: false,
				RegistryDomain:                    "quay.io",
				WorkerCount:                       3,
			},
			releaseExists:      true,
			expectedValuesJSON: alreadyMigratedJSON,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			values, err := ingressControllerValues(tc.configMapValues, tc.releaseExists)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			equalValues, err := compareValuesJSON(tc.expectedValuesJSON, values)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			if !equalValues {
				t.Fatalf("expected JSON: \n %s \n got JSON: \n %s", tc.expectedValuesJSON, values)
			}
		})
	}
}

func Test_ConfigMap_setIngressControllerTempReplicas(t *testing.T) {
	testCases := []struct {
		name                 string
		workerCount          int
		expectedTempReplicas int
		errorMatcher         func(error) bool
	}{
		{
			name:                 "case 0: basic match",
			workerCount:          3,
			expectedTempReplicas: 2,
		},
		{
			name:                 "case 1: single node",
			workerCount:          1,
			expectedTempReplicas: 1,
		},
		{
			name:                 "case 2: large cluster",
			workerCount:          20,
			expectedTempReplicas: 10,
		},
		{
			name:                 "case 3: larger cluster",
			workerCount:          50,
			expectedTempReplicas: 25,
		},
		{
			name:         "case 4: 0 workers",
			workerCount:  0,
			errorMatcher: IsInvalidExecution,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempReplicas, err := setIngressControllerTempReplicas(tc.workerCount)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if tempReplicas != tc.expectedTempReplicas {
				t.Fatal("expected", tc.expectedTempReplicas, "got", tempReplicas)
			}
		})
	}
}

func compareJSON(expectedJSON, valuesJSON string) (bool, error) {
	var err error

	expectedValues := make(map[string]interface{})
	err = json.Unmarshal([]byte(expectedJSON), &expectedValues)
	if err != nil {
		return false, microerror.Mask(err)
	}

	values := make(map[string]interface{})
	err = json.Unmarshal([]byte(valuesJSON), &values)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return reflect.DeepEqual(expectedValues, values), nil
}

func compareValuesJSON(expectedJSON string, values []byte) (bool, error) {
	var err error

	expectedValuesMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(expectedJSON), &expectedValuesMap)
	if err != nil {
		return false, microerror.Mask(err)
	}

	valuesMap := make(map[string]interface{})
	err = json.Unmarshal(values, &valuesMap)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return reflect.DeepEqual(expectedValuesMap, valuesMap), nil
}
