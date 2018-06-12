package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgofake "k8s.io/client-go/kubernetes/fake"
)

func Test_ChartConfig_newDeleteChangeForDeletePatch(t *testing.T) {
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
			description: "case 1: non-empty current and empty desired, expected current",
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
			description:  "case 2: empty current and non-empty desired, expected empty",
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
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description: "case 3: equal non-empty current and desired, expected current",
			currentState: []*v1alpha1.ChartConfig{
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
							Name: "test-chart-2",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 4: unequal non-empty current and desired, expected current",
			currentState: []*v1alpha1.ChartConfig{
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
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart-4",
						},
					},
				},
			},
			desiredState: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart-1",
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
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{
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
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart-4",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 5: incorrect type for current, expected error",
			currentState: []v1alpha1.ChartConfig{
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
			result, err := newResource.newDeleteChangeForDeletePatch(context.TODO(), tc.obj, tc.currentState, tc.desiredState)

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

func Test_ChartConfig_newDeleteChangeForUpdatePatch(t *testing.T) {
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
			description:  "case 1: empty current and non-empty desired, expected empty",
			currentState: []*v1alpha1.ChartConfig{},
			desiredState: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description: "case 2: non-empty current and empty desired, expected current",
			currentState: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			desiredState: []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 3: non-equal current and desired, expected missing current",
			currentState: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart-2",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart-3",
					},
				},
			},
			desiredState: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart-2",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart-3",
					},
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 4: incorrect type for current, expected error",
			currentState: []v1alpha1.ChartConfig{
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
			errorMatcher:         IsWrongType,
		},
		{
			description:  "case 5: incorrect type for desired, expected error",
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
			result, err := newResource.newDeleteChangeForUpdatePatch(context.TODO(), tc.obj, tc.currentState, tc.desiredState)

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
