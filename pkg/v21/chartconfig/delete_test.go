package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ChartConfig_newDeleteChangeForUpdatePatch(t *testing.T) {
	testCases := []struct {
		name                 string
		currentChartConfigs  []*v1alpha1.ChartConfig
		desiredChartConfigs  []*v1alpha1.ChartConfig
		expectedChartConfigs []*v1alpha1.ChartConfig
		errorMatcher         func(error) bool
	}{
		{
			name:                 "case 0: empty current and desired, expected empty",
			currentChartConfigs:  []*v1alpha1.ChartConfig{},
			desiredChartConfigs:  []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			name:                "case 1: empty current and non-empty desired, expected empty",
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
			name: "case 2: non-empty current and empty desired, expected current",
			currentChartConfigs: []*v1alpha1.ChartConfig{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			desiredChartConfigs: []*v1alpha1.ChartConfig{},
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
			name: "case 3: non-equal current and desired, expected missing current",
			currentChartConfigs: []*v1alpha1.ChartConfig{
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
	}

	ctx := context.Background()

	c := Config{
		Logger: microloggertest.New(),

		Provider: "aws",
	}
	cc, err := New(c)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := cc.newDeleteChangeForUpdatePatch(ctx, tc.currentChartConfigs, tc.desiredChartConfigs)

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
