package configmap

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				IngressControllerServiceEnabled:   true,
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
						"values.json": "{\"controller\":{\"replicas\":3,\"image\":{\"registry\":\"quay.io\"},\"service\":{\"enabled\":true}},\"global\":{\"controller\":{\"replicas\":3},\"migration\":{\"job\":{\"enabled\":true}}}}",
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
				IngressControllerMigrationEnabled: true,
				IngressControllerServiceEnabled:   true,
				Organization:                      "giantswarm",
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
						"values.json": "{\"controller\":{\"replicas\":7,\"image\":{\"registry\":\"quay.io\"},\"service\":{\"enabled\":true}},\"global\":{\"controller\":{\"replicas\":7},\"migration\":{\"job\":{\"enabled\":true}}}}",
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
				IngressControllerMigrationEnabled: true,
				IngressControllerServiceEnabled:   false,
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
						"values.json": "{\"controller\":{\"replicas\":3,\"image\":{\"registry\":\"quay.io\"},\"service\":{\"enabled\":false}},\"global\":{\"controller\":{\"replicas\":3},\"migration\":{\"job\":{\"enabled\":true}}}}",
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
					t.Fatalf("expected config map %s not found", expectedConfigMap.Name)
				} else if err != nil {
					t.Fatalf("expected nil, got %#v", err)
				}

				if !reflect.DeepEqual(configMap.Data, expectedConfigMap.Data) {
					t.Fatalf("expected config map data %#v, got %#v", expectedConfigMap.Data, configMap.Data)
				}

				if !reflect.DeepEqual(configMap.ObjectMeta.Labels, expectedConfigMap.ObjectMeta.Labels) {
					t.Fatalf("expected config map labels %#v, got %#v", expectedConfigMap.ObjectMeta.Labels, configMap.ObjectMeta.Labels)
				}
			}
		})
	}
}
