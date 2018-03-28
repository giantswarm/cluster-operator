package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
)

func Test_GetDesiredState(t *testing.T) {
	testCases := []struct {
		name           string
		obj            interface{}
		expectedLabels map[string]string
		expectedSpec   v1alpha1.ChartConfigSpec
		expectedType   apimetav1.TypeMeta
	}{
		{
			name: "basic match",
			obj: &v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			expectedLabels: map[string]string{
				"giantswarm.io/cluster":      "5xchu",
				"giantswarm.io/managed-by":   "cluster-operator",
				"giantswarm.io/organization": "giantswarm",
			},
			expectedSpec: v1alpha1.ChartConfigSpec{
				Chart: v1alpha1.ChartConfigSpecChart{
					Name:    "quay.io/giantswarm/chart-operator-chart",
					Channel: "stable",
					Release: "chart-operator",
				},
				VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
					Version: "0.1.0",
				},
			},
			expectedType: apimetav1.TypeMeta{
				Kind:       "ChartConfig",
				APIVersion: "core.giantswarm.io",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				BaseClusterConfig: &cluster.Config{},
				G8sClient:         fake.NewSimpleClientset(),
				K8sClient:         clientgofake.NewSimpleClientset(),
				Logger:            microloggertest.New(),
				ProjectName:       "cluster-operator",
				ToClusterGuestConfigFunc: func(v interface{}) (*v1alpha1.ClusterGuestConfig, error) {
					return v.(*v1alpha1.ClusterGuestConfig), nil
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

			if len(chartConfigs) != 1 {
				t.Fatal("expected", 1, "got", len(chartConfigs))
			}

			if !reflect.DeepEqual(chartConfigs[0].ObjectMeta.Labels, tc.expectedLabels) {
				t.Fatal("expected", tc.expectedLabels, "got", chartConfigs[0].ObjectMeta.Labels)
			}

			if !reflect.DeepEqual(chartConfigs[0].Spec, tc.expectedSpec) {
				t.Fatal("expected", tc.expectedSpec, "got", chartConfigs[0].Spec)
			}

			if !reflect.DeepEqual(chartConfigs[0].TypeMeta, tc.expectedType) {
				t.Fatal("expected", tc.expectedType, "got", chartConfigs[0].TypeMeta)
			}
		})
	}
}
