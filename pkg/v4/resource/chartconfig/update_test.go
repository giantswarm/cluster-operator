package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
)

func Test_ChartConfig_newUpdateChange(t *testing.T) {
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
			description: "case 1: non-empty current and empty desired, expected empty",
			currentState: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			desiredState:         []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description:  "case 2: empty current and noo-empty desired, expected empty",
			currentState: []*v1alpha1.ChartConfig{},
			desiredState: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description: "case 3: equal current and desired, expected empty",
			currentState: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			desiredState: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description: "case 4: unequal spec, expected desired",
			currentState: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
						},
					},
				},
			},
			desiredState: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.2-beta",
						},
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.2-beta",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 5: unequal labels, expected desired",
			currentState: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
						Labels: map[string]string{
							"cluster": "test-cluster",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
						},
					},
				},
			},
			desiredState: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
						Labels: map[string]string{
							"cluster": "test-cluster",
							"extra":   "label",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
						},
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
						Labels: map[string]string{
							"cluster": "test-cluster",
							"extra":   "label",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 6: unequal current and desired, expected changed desired",
			currentState: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
						Labels: map[string]string{
							"cluster": "test-cluster",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
						},
					},
				},
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart-2",
						Labels: map[string]string{
							"cluster": "test-cluster",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
						},
					},
				},
			},
			desiredState: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
						Labels: map[string]string{
							"cluster": "test-cluster",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.1-beta",
						},
					},
				},
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart-2",
						Labels: map[string]string{
							"cluster": "test-cluster",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.2-beta",
						},
					},
				},
			},
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				&v1alpha1.ChartConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart-2",
						Labels: map[string]string{
							"cluster": "test-cluster",
						},
					},
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name:    "test-chart",
							Channel: "0.2-beta",
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 7: incorrect type for current, expected error",
			currentState: []v1alpha1.ChartConfig{
				v1alpha1.ChartConfig{
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
			description:  "case 8: incorrect type for desired, expected error",
			currentState: []*v1alpha1.ChartConfig{},
			desiredState: []v1alpha1.ChartConfig{
				v1alpha1.ChartConfig{
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
		Provider:    label.ProviderAWS,
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
			result, err := newResource.newUpdateChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)

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
