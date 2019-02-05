package configmap

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/tenantcluster/tenantclustertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ConfigMap_newCreateChange(t *testing.T) {
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
		},
		{
			name: "case 1: non-empty current, empty desired, expected empty",
			currentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-configmap",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			desiredConfigMaps:  []*corev1.ConfigMap{},
			expectedConfigMaps: []*corev1.ConfigMap{},
		},
		{
			name: "case 2: equal non-empty current and desired, expected empty",
			currentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-configmap",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			desiredConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-configmap",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{},
		},
		{
			name:              "case 3: empty current and non-empty desired, expected empty",
			currentConfigMaps: []*corev1.ConfigMap{},
			desiredConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-configmap",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "another-configmap",
					},
					Data: map[string]string{
						"another": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-configmap",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "another-configmap",
					},
					Data: map[string]string{
						"another": "test",
					},
				},
			},
		},
		{
			name: "case 4: different non-empty current and desired, expected desired",
			currentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-configmap",
					},
					Data: map[string]string{
						"my": "test",
					},
				},
			},
			desiredConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-configmap",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "another-configmap",
					},
					Data: map[string]string{
						"another": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-configmap",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "another-configmap",
					},
					Data: map[string]string{
						"another": "test",
					},
				},
			},
		},
	}

	c := Config{
		Logger: microloggertest.New(),
		Tenant: tenantclustertest.New(tenantclustertest.Config{}),

		ProjectName: "cluster-operator",
	}
	newService, err := New(c)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := newService.newCreateChange(context.TODO(), tc.currentConfigMaps, tc.desiredConfigMaps)
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
