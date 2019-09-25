package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			description: "case 2: equal non-empty current and desired, expected empty",
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-chart",
						Namespace: resourceNamespace,
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
			description:         "case 3: empty current and non-empty desired, expected desired",
			currentChartConfigs: []*v1alpha1.ChartConfig{},
			desiredChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
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
			description: "case 4: different non-empty current and desired, expected desired",
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-chart-2",
						Namespace: resourceNamespace,
					},
				},
			},
			desiredChartConfigs: []*v1alpha1.ChartConfig{
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
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
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
			// Namespace is not compared against desired state, but pkg level variable.
			description: "case 5: different namespace current and desired, expected desired",
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-chart",
						Namespace: "foo",
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
			expectedChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
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
