package chartconfig

import (
	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_filterChartOperatorAnnotations(t *testing.T) {
	testCases := []struct {
		name               string
		chartConfig        *v1alpha1.ChartConfig
		expectedAnnotation map[string]string
	}{
		{
			name: "case 1: filter non chart-operator annotations",
			chartConfig: &v1alpha1.ChartConfig{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/mananged-by":  "giantswarm",
						"chart-operator.giantswarm.io/organization": "giantswarm",
						"random-generated-by":                       "kubeconfig",
					},
				},
			},
			expectedAnnotation: map[string]string{
				"chart-operator.giantswarm.io/mananged-by":  "giantswarm",
				"chart-operator.giantswarm.io/organization": "giantswarm",
			},
		},
		{
			name: "case 2: skip comparing cordon annotations",
			chartConfig: &v1alpha1.ChartConfig{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/mananged-by":  "giantswarm",
						"chart-operator.giantswarm.io/organization": "giantswarm",
						annotation.CordonReason:                     "managing cordns",
						annotation.CordonUntilDate:                  "2019-12-31T23:59:59Z",
					},
				},
			},
			expectedAnnotation: map[string]string{
				"chart-operator.giantswarm.io/mananged-by":  "giantswarm",
				"chart-operator.giantswarm.io/organization": "giantswarm",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filterChartOperatorAnnotations(tc.chartConfig)

			if !reflect.DeepEqual(result, tc.expectedAnnotation) {
				t.Fatalf("expected annotation %#q, got %#q", tc.expectedAnnotation, result)
			}
		})
	}
}
