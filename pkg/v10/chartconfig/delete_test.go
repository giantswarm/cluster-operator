package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/tenantcluster/tenantclustertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ChartConfig_newDeleteChangeForDeletePatch(t *testing.T) {
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
			description: "case 1: non-empty current and empty desired, expected current",
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
			description:         "case 2: empty current and non-empty desired, expected empty",
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
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
			errorMatcher:         nil,
		},
		{
			description: "case 3: equal non-empty current and desired, expected current",
			currentChartConfigs: []*v1alpha1.ChartConfig{
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
			currentChartConfigs: []*v1alpha1.ChartConfig{
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
			desiredChartConfigs: []*v1alpha1.ChartConfig{
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
	}

	c := Config{
		Logger: microloggertest.New(),
		Tenant: tenantclustertest.New(tenantclustertest.Config{}),

		ProjectName: "cluster-operator",
	}
	cc, err := New(c)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := cc.newDeleteChangeForDeletePatch(context.TODO(), tc.currentChartConfigs, tc.desiredChartConfigs)

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
			description:         "case 1: empty current and non-empty desired, expected empty",
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
			description: "case 2: non-empty current and empty desired, expected current",
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
			description: "case 3: non-equal current and desired, expected missing current",
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

	c := Config{
		Logger: microloggertest.New(),
		Tenant: tenantclustertest.New(tenantclustertest.Config{}),

		ProjectName: "cluster-operator",
	}
	cc, err := New(c)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := cc.newDeleteChangeForUpdatePatch(context.TODO(), tc.currentChartConfigs, tc.desiredChartConfigs)

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
