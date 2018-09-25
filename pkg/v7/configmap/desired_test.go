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
		name                   string
		clusterConfig          ClusterConfig
		configMapValues        ConfigMapValues
		providerChartSpecs     []key.ChartSpec
		expectedConfigMapSpecs []ConfigMapSpec
	}{
		{
			name: "case 0: basic match",
			clusterConfig: ClusterConfig{
				APIDomain:  "5xchu.aws.giantswarm.io",
				ClusterID:  "5xchu",
				Namespaces: []string{},
			},
			configMapValues: ConfigMapValues{
				ClusterID:                         "5xchu",
				IngressControllerMigrationEnabled: true,
				IngressControllerUseProxyProtocol: true,
				Organization:                      "giantswarm",
				RegistryDomain:                    "quay.io",
				WorkerCount:                       3,
			},
			expectedConfigMapSpecs: []ConfigMapSpec{
				{
					Name:      "coredns-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "cert-exporter-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "kube-state-metrics-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "net-exporter-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "nginx-ingress-controller-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "node-exporter-values",
					Namespace: metav1.NamespaceSystem,
				},
			},
		},
		{
			name: "case 1: provider chart without configmap",
			clusterConfig: ClusterConfig{
				APIDomain:  "5xchu.aws.giantswarm.io",
				ClusterID:  "5xchu",
				Namespaces: []string{},
			},
			configMapValues: ConfigMapValues{
				ClusterID:                         "5xchu",
				Organization:                      "giantswarm",
				IngressControllerMigrationEnabled: true,
				IngressControllerUseProxyProtocol: true,
				RegistryDomain:                    "quay.io",
				WorkerCount:                       7,
			},
			providerChartSpecs: []key.ChartSpec{
				{
					AppName:   "test-app",
					ChartName: "test-app-chart",
					Namespace: metav1.NamespaceSystem,
				},
			},
			expectedConfigMapSpecs: []ConfigMapSpec{
				{
					Name:      "coredns-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "cert-exporter-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "kube-state-metrics-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "net-exporter-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "nginx-ingress-controller-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "node-exporter-values",
					Namespace: metav1.NamespaceSystem,
				},
			},
		},
		{
			name: "case 2: provider chart with configmap in different namespace",
			clusterConfig: ClusterConfig{
				APIDomain:  "5xchu.aws.giantswarm.io",
				ClusterID:  "5xchu",
				Namespaces: []string{},
			},
			configMapValues: ConfigMapValues{
				ClusterID:                         "5xchu",
				Organization:                      "giantswarm",
				IngressControllerMigrationEnabled: true,
				IngressControllerUseProxyProtocol: true,
				RegistryDomain:                    "quay.io",
				WorkerCount:                       7,
			},
			providerChartSpecs: []key.ChartSpec{
				{
					AppName:       "test-app",
					ChartName:     "test-app-chart",
					ConfigMapName: "test-app-values",
					Namespace:     "giantswarm",
				},
			},
			expectedConfigMapSpecs: []ConfigMapSpec{
				{
					Name:      "coredns-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "cert-exporter-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "kube-state-metrics-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "net-exporter-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "nginx-ingress-controller-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "node-exporter-values",
					Namespace: metav1.NamespaceSystem,
				},
				{
					Name:      "test-app-values",
					Namespace: "giantswarm",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				Logger: microloggertest.New(),
				Tenant: &tenantMock{
					fakeTenantHelmClient: &helmMock{},
				},

				ProjectName: "cluster-operator",
			}
			newService, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			configMaps, err := newService.GetDesiredState(context.TODO(), tc.clusterConfig, tc.configMapValues, tc.providerChartSpecs)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if len(configMaps) != len(tc.expectedConfigMapSpecs) {
				t.Fatal("expected", len(tc.expectedConfigMapSpecs), "got", len(configMaps))
			}

			for _, expectedSpec := range tc.expectedConfigMapSpecs {
				_, err := getConfigMapByNameAndNamespace(configMaps, expectedSpec.Name, expectedSpec.Namespace)
				if IsNotFound(err) {
					t.Fatalf("expected chart %#q/%#q not found", expectedSpec.Namespace, expectedSpec.Name)
				} else if err != nil {
					t.Fatalf("expected nil, got %#v", err)
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
