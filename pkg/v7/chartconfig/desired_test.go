package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/v7/key"
)

func Test_ChartConfig_GetDesiredState(t *testing.T) {
	commonChartConfigNames := []string{
		"cert-exporter-chart",
		"kubernetes-kube-state-metrics-chart",
		"kubernetes-nginx-ingress-controller-chart",
		"kubernetes-node-exporter-chart",
		"net-exporter-chart",
	}

	testCases := []struct {
		name                             string
		clusterConfig                    ClusterConfig
		providerChartSpecs               []key.ChartSpec
		expectedProviderChartConfigNames []string
	}{
		{
			name: "case 0: basic match",
			clusterConfig: ClusterConfig{
				APIDomain:    "api.5xchu.aws.giantswarm.io",
				ClusterID:    "5xchu",
				Organization: "giantswarm",
			},
			expectedProviderChartConfigNames: []string{},
		},
		{
			name: "case 1: single provider chart",
			clusterConfig: ClusterConfig{
				APIDomain:    "api.eggs2.azure.giantswarm.io",
				ClusterID:    "eggs2",
				Organization: "giantswarm",
			},
			providerChartSpecs: []key.ChartSpec{
				{
					AppName:     "test-app",
					ChannelName: "0-1-stable",
					ChartName:   "test-app-chart",
					Namespace:   metav1.NamespaceSystem,
					ReleaseName: "test-app",
				},
			},
			expectedProviderChartConfigNames: []string{
				"test-app-chart",
			},
		},
		{
			name: "case 2: multiple provider charts",
			clusterConfig: ClusterConfig{
				APIDomain:    "api.eggs2.azure.giantswarm.io",
				ClusterID:    "eggs2",
				Organization: "giantswarm",
			},
			providerChartSpecs: []key.ChartSpec{
				{
					AppName:       "test-app",
					ChannelName:   "0-1-stable",
					ChartName:     "test-app-chart",
					ConfigMapName: "test-app-values",
					Namespace:     metav1.NamespaceSystem,
					ReleaseName:   "test-app",
				},
				{
					AppName:       "test-app2",
					ChannelName:   "0-1-stable",
					ChartName:     "test-app2-chart",
					ConfigMapName: "test-app2-values",
					Namespace:     metav1.NamespaceSystem,
					ReleaseName:   "test-app2",
				},
			},
			expectedProviderChartConfigNames: []string{
				"test-app-chart",
				"test-app2-chart",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeTenantK8sClient := clientgofake.NewSimpleClientset()

			c := Config{
				Logger: microloggertest.New(),
				Tenant: &tenantMock{
					fakeTenantK8sClient: fakeTenantK8sClient,
				},

				ProjectName: "cluster-operator",
			}
			cc, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			chartConfigs, err := cc.GetDesiredState(context.TODO(), tc.clusterConfig, tc.providerChartSpecs)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			expectedChartConfigNames := append(commonChartConfigNames, tc.expectedProviderChartConfigNames...)
			if len(chartConfigs) != len(expectedChartConfigNames) {
				t.Fatal("expected", len(expectedChartConfigNames), "got", len(chartConfigs))
			}

			for _, expectedName := range expectedChartConfigNames {
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

func Test_ChartConfig_newConfigMapSpec(t *testing.T) {
	testCases := []struct {
		name                  string
		clusterConfig         ClusterConfig
		configMapName         string
		namespace             string
		presentConfigMaps     []*corev1.ConfigMap
		expectedConfigMapSpec *v1alpha1.ChartConfigSpecConfigMap
	}{
		{
			name: "case 0: basic match with no configmaps",
			clusterConfig: ClusterConfig{
				ClusterID: "5xchu",
			},
			configMapName:     "ingress-controller-values",
			namespace:         metav1.NamespaceSystem,
			presentConfigMaps: []*corev1.ConfigMap{},
			expectedConfigMapSpec: &v1alpha1.ChartConfigSpecConfigMap{
				Name:      "ingress-controller-values",
				Namespace: metav1.NamespaceSystem,
			},
		},
		{
			name: "case 1: no matching configmaps",
			clusterConfig: ClusterConfig{
				ClusterID: "5xchu",
			},
			configMapName: "ingress-controller-values",
			namespace:     metav1.NamespaceSystem,
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
			clusterConfig: ClusterConfig{
				ClusterID: "5xchu",
			},
			configMapName: "ingress-controller-values",
			namespace:     metav1.NamespaceSystem,
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
			clusterConfig: ClusterConfig{
				ClusterID: "5xchu",
			},
			configMapName: "ingress-controller-values",
			namespace:     metav1.NamespaceSystem,
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
			clusterConfig: ClusterConfig{
				ClusterID: "5xchu",
			},
			configMapName: "ingress-controller-values",
			namespace:     metav1.NamespaceSystem,
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

			fakeTenantK8sClient := clientgofake.NewSimpleClientset(objs...)

			c := Config{
				Logger: microloggertest.New(),
				Tenant: &tenantMock{
					fakeTenantK8sClient: fakeTenantK8sClient,
				},

				ProjectName: "cluster-operator",
			}
			cc, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := cc.newConfigMapSpec(context.TODO(), tc.clusterConfig, tc.configMapName, tc.namespace)
			if err != nil {
				t.Fatalf("expected nil, got %#v", err)
			}

			if !reflect.DeepEqual(result, tc.expectedConfigMapSpec) {
				t.Fatalf("expected config map spec %#v, got %#v", tc.expectedConfigMapSpec, result)
			}
		})
	}
}
