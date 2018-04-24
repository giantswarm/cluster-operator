package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
)

func Test_ChartConfig_GetDesiredState(t *testing.T) {
	testCases := []struct {
		name               string
		obj                interface{}
		expectedChartNames []string
		expectedLabels     map[string]string
	}{
		{
			name: "basic match",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			expectedChartNames: []string{
				"kubernetes-kube-state-metrics-chart",
			},
			expectedLabels: map[string]string{
				"giantswarm.io/cluster":      "5xchu",
				"giantswarm.io/managed-by":   "cluster-operator",
				"giantswarm.io/organization": "giantswarm",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				BaseClusterConfig: cluster.Config{
					ClusterID: "test-cluster",
				},
				G8sClient:   fake.NewSimpleClientset(),
				Guest:       &guestMock{},
				K8sClient:   clientgofake.NewSimpleClientset(),
				Logger:      microloggertest.New(),
				ProjectName: "cluster-operator",
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

			if len(chartConfigs) != len(tc.expectedChartNames) {
				t.Fatal("expected", len(tc.expectedChartNames), "got", len(chartConfigs))
			}

			for _, chartName := range tc.expectedChartNames {
				chart, err := getChartConfigByName(chartConfigs, chartName)
				if IsNotFound(err) {
					t.Fatalf("expected chart '%s' got not found error", chartName)
				} else if err != nil {
					t.Fatal("expected", nil, "got", err)
				}

				if !reflect.DeepEqual(chart.Labels, tc.expectedLabels) {
					t.Fatalf("expected labels '%q' got '%q'", tc.expectedLabels, chart.Labels)
				}
			}
		})
	}
}
