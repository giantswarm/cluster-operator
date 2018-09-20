package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
)

func Test_ChartConfig_GetDesiredState(t *testing.T) {
	testCases := []struct {
		name                     string
		obj                      interface{}
		provider                 string
		expectedChartConfigNames []string
	}{
		{
			name: "basic match for aws",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			provider: label.ProviderAWS,
			expectedChartConfigNames: []string{
				"cert-exporter-chart",
				// "kubernetes-coredns-chart",
				"kubernetes-kube-state-metrics-chart",
				"kubernetes-nginx-ingress-controller-chart",
				"kubernetes-node-exporter-chart",
				"net-exporter-chart",
			},
		},
		{
			name: "basic match for kvm",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.kvm.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			provider: label.ProviderKVM,
			expectedChartConfigNames: []string{
				"cert-exporter-chart",
				// "kubernetes-coredns-chart",
				"kubernetes-kube-state-metrics-chart",
				"kubernetes-nginx-ingress-controller-chart",
				"kubernetes-node-exporter-chart",
				"net-exporter-chart",
			},
		},
		{
			name: "azure also has provider specific chartconfigs",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.azure.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			provider: label.ProviderAzure,
			expectedChartConfigNames: []string{
				"cert-exporter-chart",
				// "kubernetes-coredns-chart",
				"kubernetes-external-dns-chart",
				"kubernetes-kube-state-metrics-chart",
				"kubernetes-nginx-ingress-controller-chart",
				"kubernetes-node-exporter-chart",
				"net-exporter-chart",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				BaseClusterConfig: cluster.Config{
					ClusterID: "test-cluster",
				},
				G8sClient: fake.NewSimpleClientset(),
				Guest: &guestMock{
					fakeGuestK8sClient: clientgofake.NewSimpleClientset(),
				},
				K8sClient:   clientgofake.NewSimpleClientset(),
				Logger:      microloggertest.New(),
				ProjectName: "cluster-operator",
				Provider:    tc.provider,
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
			}
			newResource, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := newResource.GetDesiredState(context.TODO(), tc.obj)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			chartConfigs, err := toChartConfigs(result)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if len(chartConfigs) != len(tc.expectedChartConfigNames) {
				t.Fatal("expected", len(tc.expectedChartConfigNames), "got", len(chartConfigs))
			}

			for _, expectedName := range tc.expectedChartConfigNames {
				_, err := getChartConfigByName(chartConfigs, expectedName)
				if IsNotFound(err) {
					t.Fatalf("expected chart %s not found", expectedName)
				} else if err != nil {
					t.Fatalf("expected nil, got %#v", err)
				}
			}
		})
	}
}

func Test_ChartConfig_getConfigMapSpec(t *testing.T) {
	testCases := []struct {
		name                  string
		clusterConfig         cluster.Config
		configMapName         string
		configMapNamespace    string
		presentConfigMaps     []*corev1.ConfigMap
		expectedConfigMapSpec *v1alpha1.ChartConfigSpecConfigMap
	}{
		{
			name: "case 0: basic match with no configmaps",
			clusterConfig: cluster.Config{
				ClusterID: "5xchu",
			},
			configMapName:      "ingress-controller-values",
			configMapNamespace: metav1.NamespaceSystem,
			presentConfigMaps:  []*corev1.ConfigMap{},
			expectedConfigMapSpec: &v1alpha1.ChartConfigSpecConfigMap{
				Name:      "ingress-controller-values",
				Namespace: metav1.NamespaceSystem,
			},
		},
		{
			name: "case 1: no matching configmaps",
			clusterConfig: cluster.Config{
				ClusterID: "5xchu",
			},
			configMapName:      "ingress-controller-values",
			configMapNamespace: metav1.NamespaceSystem,
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"key": "value",
					},
				},
			},
			expectedConfigMapSpec: &v1alpha1.ChartConfigSpecConfigMap{
				Name:      "ingress-controller-values",
				Namespace: metav1.NamespaceSystem,
			},
		},
		{
			name: "case 2: configmap in different namespace",
			clusterConfig: cluster.Config{
				ClusterID: "5xchu",
			},
			configMapName:      "ingress-controller-values",
			configMapNamespace: metav1.NamespaceSystem,
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "ingress-controller-values",
						Namespace:       metav1.NamespacePublic,
						ResourceVersion: "12345",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
			},
			expectedConfigMapSpec: &v1alpha1.ChartConfigSpecConfigMap{
				Name:      "ingress-controller-values",
				Namespace: metav1.NamespaceSystem,
			},
		},
		{
			name: "case 3: matching configmap, resource version is set",
			clusterConfig: cluster.Config{
				ClusterID: "5xchu",
			},
			configMapName:      "ingress-controller-values",
			configMapNamespace: metav1.NamespaceSystem,
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "ingress-controller-values",
						Namespace:       metav1.NamespaceSystem,
						ResourceVersion: "12345",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
			},
			expectedConfigMapSpec: &v1alpha1.ChartConfigSpecConfigMap{
				Name:            "ingress-controller-values",
				Namespace:       metav1.NamespaceSystem,
				ResourceVersion: "12345",
			},
		},
		{
			name: "case 4: multiple configmaps, correct resource version is set",
			clusterConfig: cluster.Config{
				ClusterID: "5xchu",
			},
			configMapName:      "ingress-controller-values",
			configMapNamespace: metav1.NamespaceSystem,
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "ingress-controller-values",
						Namespace:       metav1.NamespaceSystem,
						ResourceVersion: "12345",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-values",
						Namespace:       metav1.NamespaceSystem,
						ResourceVersion: "67890",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
			},
			expectedConfigMapSpec: &v1alpha1.ChartConfigSpecConfigMap{
				Name:            "ingress-controller-values",
				Namespace:       metav1.NamespaceSystem,
				ResourceVersion: "12345",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tc.presentConfigMaps))
			for _, cm := range tc.presentConfigMaps {
				objs = append(objs, cm)
			}

			fakeGuestK8sClient := clientgofake.NewSimpleClientset(objs...)
			guestService := &guestMock{
				fakeGuestK8sClient: fakeGuestK8sClient,
			}

			c := Config{
				BaseClusterConfig: cluster.Config{
					ClusterID: "test-cluster",
				},
				G8sClient:   fake.NewSimpleClientset(),
				Guest:       guestService,
				K8sClient:   clientgofake.NewSimpleClientset(),
				Logger:      microloggertest.New(),
				ProjectName: "cluster-operator",
				Provider:    label.ProviderAWS,
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
			}
			newResource, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := newResource.getConfigMapSpec(context.TODO(), tc.clusterConfig, tc.configMapName, tc.configMapNamespace)
			if err != nil {
				t.Fatalf("expected nil, got %#v", err)
			}

			if !reflect.DeepEqual(result, tc.expectedConfigMapSpec) {
				t.Fatalf("expected config map spec %#v, got %#v", tc.expectedConfigMapSpec, result)
			}
		})
	}
}
