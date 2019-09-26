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

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v21/key"
)

func Test_ChartConfig_GetDesiredState(t *testing.T) {
	commonChartConfigNames := []string{
		"cert-exporter-chart",
		"kubernetes-coredns-chart",
		"kubernetes-kube-state-metrics-chart",
		"kubernetes-metrics-server-chart",
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

				Provider: "aws",
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

func Test_ChartConfig_newChartConfig(t *testing.T) {
	testCases := []struct {
		name                string
		clusterConfig       ClusterConfig
		chartSpec           key.ChartSpec
		expectedChartConfig *v1alpha1.ChartConfig
	}{
		{
			name: "case 0: basic match",
			clusterConfig: ClusterConfig{
				ClusterID:    "5xchu",
				Organization: "giantswarm",
			},
			chartSpec: key.ChartSpec{
				AppName:       "kube-state-metrics",
				ChannelName:   "0-1-stable",
				ChartName:     "kubernetes-kube-state-metrics-chart",
				ConfigMapName: "kube-state-metrics-values",
				Namespace:     metav1.NamespaceSystem,
				ReleaseName:   "kube-state-metrics",
			},
			expectedChartConfig: &v1alpha1.ChartConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kubernetes-kube-state-metrics-chart",
					Labels: map[string]string{
						label.App:          "kube-state-metrics",
						label.Cluster:      "5xchu",
						label.ManagedBy:    "cluster-operator",
						label.Organization: "giantswarm",
						label.ServiceType:  label.ServiceTypeManaged,
					},
				},
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:      "kubernetes-kube-state-metrics-chart",
						Namespace: metav1.NamespaceSystem,
						Channel:   "0-1-stable",
						Release:   "kube-state-metrics",

						ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:            "kube-state-metrics-values",
							Namespace:       metav1.NamespaceSystem,
							ResourceVersion: "",
						},
						Secret: v1alpha1.ChartConfigSpecSecret{
							Name:            "",
							Namespace:       "",
							ResourceVersion: "",
						},
						UserConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:            "",
							Namespace:       "",
							ResourceVersion: "",
						},
					},
					VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
						Version: chartConfigVersionBundleVersion,
					},
				},
			},
		},
		{
			name: "case 1: basic match with user configmap",
			clusterConfig: ClusterConfig{
				ClusterID:    "5xchu",
				Organization: "giantswarm",
			},
			chartSpec: key.ChartSpec{
				AppName:           "nginx-ingress-controller",
				ChannelName:       "0-3-stable",
				ChartName:         "kubernetes-nginx-ingress-controller-chart",
				ConfigMapName:     "nginx-ingress-controller-values",
				Namespace:         metav1.NamespaceSystem,
				ReleaseName:       "nginx-ingress-controller",
				UserConfigMapName: "nginx-ingress-controller-user-values",
			},
			expectedChartConfig: &v1alpha1.ChartConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kubernetes-nginx-ingress-controller-chart",
					Labels: map[string]string{
						label.App:          "nginx-ingress-controller",
						label.Cluster:      "5xchu",
						label.ManagedBy:    "cluster-operator",
						label.Organization: "giantswarm",
						label.ServiceType:  label.ServiceTypeManaged,
					},
				},
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:      "kubernetes-nginx-ingress-controller-chart",
						Namespace: metav1.NamespaceSystem,
						Channel:   "0-3-stable",
						Release:   "nginx-ingress-controller",

						ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:            "nginx-ingress-controller-values",
							Namespace:       metav1.NamespaceSystem,
							ResourceVersion: "",
						},
						Secret: v1alpha1.ChartConfigSpecSecret{
							Name:            "",
							Namespace:       "",
							ResourceVersion: "",
						},
						UserConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:            "nginx-ingress-controller-user-values",
							Namespace:       metav1.NamespaceSystem,
							ResourceVersion: "",
						},
					},
					VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
						Version: chartConfigVersionBundleVersion,
					},
				},
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

				Provider: "aws",
			}
			cc, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := cc.newChartConfig(context.TODO(), tc.clusterConfig, tc.chartSpec)
			if err != nil {
				t.Fatalf("expected nil, got %#v", err)
			}

			if result.Name != tc.expectedChartConfig.Name {
				t.Fatalf("expected chart config name %#q, got %#q", tc.expectedChartConfig.Name, result.Name)
			}

			if !reflect.DeepEqual(result.Spec, tc.expectedChartConfig.Spec) {
				t.Fatalf("expected chart config spec %#v, got %#v", tc.expectedChartConfig.Spec, result.Spec)
			}

			if !reflect.DeepEqual(result.Labels, tc.expectedChartConfig.Labels) {
				t.Fatalf("expected chart config labels %#v, got %#v", tc.expectedChartConfig.Labels, result.Labels)
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

				Provider: "aws",
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
