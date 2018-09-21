package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_ChartConfig_newCreateChange(t *testing.T) {
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
			description: "case 1: non-empty current, empty desired, expected empty",
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
			},
			desiredChartConfigs:  []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description: "case 2: equal non-empty current and desired, expected empty",
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart",
						},
					},
				},
			},
			desiredChartConfigs: []*v1alpha1.ChartConfig{
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
			description:         "case 3: empty current and non-empty desired, expected desired",
			currentChartConfigs: []*v1alpha1.ChartConfig{},
			desiredChartConfigs: []*v1alpha1.ChartConfig{
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
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
					Spec: v1alpha1.ChartConfigSpec{
						Chart: v1alpha1.ChartConfigSpecChart{
							Name: "test-chart-2",
						},
					},
				},
			},
			desiredChartConfigs: []*v1alpha1.ChartConfig{
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
	}

	c := Config{
		Logger: microloggertest.New(),
		Tenant: &tenantMock{},

		ProjectName: "cluster-operator",
	}
	cc, err := New(c)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := cc.newCreateChange(context.TODO(), tc.currentChartConfigs, tc.desiredChartConfigs)

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
