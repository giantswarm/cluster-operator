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
)

const (
	basicMatchJSON = `
	{
		"controller": {
			"replicas": 3
		},
		"global": {
			"controller": {
				"replicas": 3,
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
			"replicas": 7
		},
		"global": {
			"controller": {
				"replicas": 7,
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
			"replicas": 3
		},
		"global": {
			"controller": {
				"replicas": 3,
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
		name               string
		configMapValues    ConfigMapValues
		expectedConfigMaps []*corev1.ConfigMap
	}{
		{
			name: "case 0: basic match",
			configMapValues: ConfigMapValues{
				ClusterID:                         "5xchu",
				IngressControllerMigrationEnabled: true,
				IngressControllerUseProxyProtocol: true,
				Organization:                      "giantswarm",
				WorkerCount:                       3,
			},
			expectedConfigMaps: []*corev1.ConfigMap{
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
			configMapValues: ConfigMapValues{
				ClusterID:                         "5xchu",
				Organization:                      "giantswarm",
				IngressControllerMigrationEnabled: true,
				IngressControllerUseProxyProtocol: true,
				WorkerCount:                       7,
			},
			expectedConfigMaps: []*corev1.ConfigMap{
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
			configMapValues: ConfigMapValues{
				ClusterID:                         "5xchu",
				IngressControllerMigrationEnabled: false,
				IngressControllerUseProxyProtocol: false,
				Organization:                      "giantswarm",
				WorkerCount:                       3,
			},
			expectedConfigMaps: []*corev1.ConfigMap{
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				Guest:          &guestMock{},
				Logger:         microloggertest.New(),
				ProjectName:    "cluster-operator",
				RegistryDomain: "quay.io",
			}
			newService, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			configMaps, err := newService.GetDesiredState(context.TODO(), tc.configMapValues)
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
