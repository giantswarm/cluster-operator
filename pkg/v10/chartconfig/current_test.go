package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_ChartConfig_GetCurrentState(t *testing.T) {
	testCases := []struct {
		name                 string
		clusterConfig        ClusterConfig
		presentChartConfigs  []*v1alpha1.ChartConfig
		expectedChartConfigs []*v1alpha1.ChartConfig
	}{
		{
			name: "case 0: no results",
			clusterConfig: ClusterConfig{
				APIDomain:    "api.5xchu.aws.giantswarm.io",
				ClusterID:    "5xchu",
				Organization: "giantswarm",
			},
			presentChartConfigs:  []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
		},
		{
			name: "case 1: no match without managed-by label",
			clusterConfig: ClusterConfig{
				APIDomain:    "api.5xchu.aws.giantswarm.io",
				ClusterID:    "5xchu",
				Organization: "giantswarm",
			},
			presentChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-chart",
						Namespace: "giantswarm",
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
							Release: "test-release",
						},
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
		},
		{
			name: "case 2: single result",
			clusterConfig: ClusterConfig{
				APIDomain:    "api.5xchu.aws.giantswarm.io",
				ClusterID:    "5xchu",
				Organization: "giantswarm",
			},
			presentChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-chart",
						Namespace: "giantswarm",
						Labels: map[string]string{
							"giantswarm.io/managed-by": "cluster-operator",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
							Release: "test-release",
						},
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-chart",
						Namespace: "giantswarm",
						Labels: map[string]string{
							"giantswarm.io/managed-by": "cluster-operator",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
							Release: "test-release",
						},
					},
				},
			},
		},
		{
			name: "case 2: multiple results",
			clusterConfig: ClusterConfig{
				APIDomain:    "api.5xchu.aws.giantswarm.io",
				ClusterID:    "5xchu",
				Organization: "giantswarm",
			},
			presentChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-chart",
						Namespace: "giantswarm",
						Labels: map[string]string{
							"giantswarm.io/managed-by": "cluster-operator",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
							Release: "test-release",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "another-chart",
						Namespace: "giantswarm",
						Labels: map[string]string{
							"giantswarm.io/managed-by": "cluster-operator",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "another-chart",
							Channel: "0.1-beta",
							Release: "another-release",
						},
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-chart",
						Namespace: "giantswarm",
						Labels: map[string]string{
							"giantswarm.io/managed-by": "cluster-operator",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
							Release: "test-release",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "another-chart",
						Namespace: "giantswarm",
						Labels: map[string]string{
							"giantswarm.io/managed-by": "cluster-operator",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "another-chart",
							Channel: "0.1-beta",
							Release: "another-release",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tc.presentChartConfigs))
			for _, cc := range tc.presentChartConfigs {
				objs = append(objs, cc)
			}

			fakeTenantG8sClient := fake.NewSimpleClientset(objs...)

			c := Config{
				Logger: microloggertest.New(),
				Tenant: &tenantMock{
					fakeTenantG8sClient: fakeTenantG8sClient,
				},

				ProjectName: "cluster-operator",
			}
			cc, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			chartConfigs, err := cc.GetCurrentState(context.TODO(), tc.clusterConfig)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if len(chartConfigs) != len(tc.expectedChartConfigs) {
				t.Fatalf("expected %d chartconfigs got %d", len(tc.expectedChartConfigs), len(chartConfigs))
			}

			for _, cc := range chartConfigs {
				found := false
				for _, ec := range tc.expectedChartConfigs {
					if reflect.DeepEqual(cc, ec) {
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("unexpected ChartConfig %#v among returned values", *cc)
				}
			}
		})
	}
}
