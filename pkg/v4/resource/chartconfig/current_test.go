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
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
)

func Test_ChartConfig_GetCurrentState(t *testing.T) {
	testCases := []struct {
		name                 string
		obj                  interface{}
		presentChartConfigs  []*v1alpha1.ChartConfig
		expectedChartConfigs []*v1alpha1.ChartConfig
	}{
		{
			name: "case 0: no results",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			presentChartConfigs:  []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
		},
		{
			name: "case 1: single result",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
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
			expectedChartConfigs: []*v1alpha1.ChartConfig{
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
		},
		{
			name: "case 2: multiple results",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
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
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "another-chart",
						Namespace: "giantswarm",
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

			fakeGuestG8sClient := fake.NewSimpleClientset(objs...)
			guestService := &guestMock{
				fakeGuestG8sClient: fakeGuestG8sClient,
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
				Provider:    "aws",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
			}
			newResource, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := newResource.GetCurrentState(context.TODO(), tc.obj)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			chartConfigs, err := toChartConfigs(result)
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
