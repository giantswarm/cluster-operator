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

func Test_ChartConfig_newCreateChange(t *testing.T) {
	testCases := []struct {
		description          string
		obj                  interface{}
		currentState         interface{}
		desiredState         interface{}
		expectedChartConfigs []*v1alpha1.ChartConfig
		errorMatcher         func(error) bool
	}{
		{
			description:          "case 0: empty current and desired, expected empty",
			currentState:         []*v1alpha1.ChartConfig{},
			desiredState:         []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description: "case 1: non-empty current, empty desired, expected empty",
			currentState: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
			},
			desiredState:         []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description: "case 2: equal non-empty current and desired, expected empty",
			currentState: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
			},
			desiredState: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description:  "case 3: empty current and non-empty desired, expected desired",
			currentState: []*v1alpha1.ChartConfig{},
			desiredState: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 4: different non-empty current and desired, expected desired",
			currentState: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart-2",
						},
					},
				},
			},
			desiredState: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart-2",
						},
					},
				},
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart-3",
						},
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart-3",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 5: incorrect type for current state",
			currentState: []v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
			},
			desiredState:         []v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         IsWrongType,
		},
		{
			description:  "case 6: incorrect type for desired state",
			currentState: []*v1alpha1.ChartConfig{},
			desiredState: []v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         IsWrongType,
		},
	}

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

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newCreateChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)

			if err != nil {
				switch {
				case err == nil && tc.errorMatcher == nil: // correct; carry on
				case err != nil && tc.errorMatcher != nil:
					if !tc.errorMatcher(err) {
						t.Fatalf("error == %#v, want matching", err)
					}
				case err != nil && tc.errorMatcher == nil:
					t.Fatalf("error == %#v, want nil", err)
				case err == nil && tc.errorMatcher != nil:
					t.Fatalf("error == nil, want non-nil")
				}
			} else if !reflect.DeepEqual(result, tc.expectedChartConfigs) {
				t.Fatalf("expected %#v got %#v", tc.expectedChartConfigs, result)
			}
		})
	}
}
