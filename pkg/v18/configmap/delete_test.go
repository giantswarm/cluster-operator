package configmap

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ConfigMap_newDeleteChangeForUpdatePatch(t *testing.T) {
	testCases := []struct {
		name               string
		currentConfigMaps  []*corev1.ConfigMap
		desiredConfigMaps  []*corev1.ConfigMap
		expectedConfigMaps []*corev1.ConfigMap
		errorMatcher       func(error) bool
	}{
		{
			name:               "case 0: empty current and desired, expected empty",
			currentConfigMaps:  []*corev1.ConfigMap{},
			desiredConfigMaps:  []*corev1.ConfigMap{},
			expectedConfigMaps: []*corev1.ConfigMap{},
			errorMatcher:       nil,
		},
		{
			name:              "case 1: empty current and non-empty desired, expected empty",
			currentConfigMaps: []*corev1.ConfigMap{},
			desiredConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{},
			errorMatcher:       nil,
		},
		{
			name: "case 2: non-empty current and empty desired, expected current",
			currentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			desiredConfigMaps: []*corev1.ConfigMap{},
			expectedConfigMaps: []*corev1.ConfigMap{
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
			currentConfigMaps: []*corev1.ConfigMap{
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
			desiredConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-chart",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
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
		Tenant: &tenantMock{},
	}
	newService, err := New(c)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := newService.newDeleteChangeForUpdatePatch(ctx, tc.currentConfigMaps, tc.desiredConfigMaps)

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
			} else if !reflect.DeepEqual(result, tc.expectedConfigMaps) {
				t.Fatalf("expected %#v got %#v", tc.expectedConfigMaps, result)
			}
		})
	}
}
