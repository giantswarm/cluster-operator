package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ChartConfig_newUpdateChange(t *testing.T) {
	testCases := []struct {
		description          string
		currentChartConfigs  []*v1alpha1.ChartConfig
		desiredChartConfigs  []*v1alpha1.ChartConfig
		expectedChartConfigs []*v1alpha1.ChartConfig
		errorMatcher         func(error) bool
	}{
		{
			description:          "case 0: empty current and desired, expected empty",
			currentChartConfigs:  []*v1alpha1.ChartConfig{},
			desiredChartConfigs:  []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description: "case 1: non-empty current and empty desired, expected empty",
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			desiredChartConfigs:  []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description:         "case 2: empty current and noo-empty desired, expected empty",
			currentChartConfigs: []*v1alpha1.ChartConfig{},
			desiredChartConfigs: []*v1alpha1.ChartConfig{
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
			description: "case 3: equal current and desired, expected empty",
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			desiredChartConfigs: []*v1alpha1.ChartConfig{
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
			description: "case 4: unequal spec, expected desired",
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
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
			desiredChartConfigs: []*v1alpha1.ChartConfig{
				{
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
				{
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
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
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
			desiredChartConfigs: []*v1alpha1.ChartConfig{
				{
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
				{
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
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
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
				{
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
			desiredChartConfigs: []*v1alpha1.ChartConfig{
				{
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
				{
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
				{
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
	}

	c := Config{
		Logger: microloggertest.New(),
		Tenant: &tenantMock{},

		Provider: "aws",
	}
	cc, err := New(c)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := cc.newUpdateChange(context.TODO(), tc.currentChartConfigs, tc.desiredChartConfigs)

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
