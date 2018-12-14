package configmap

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ConfigMap_newDeleteChangeForDeletePatch(t *testing.T) {
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
			name: "case 1: non-empty current and empty desired, expected current",
			currentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
			},
			desiredConfigMaps: []*corev1.ConfigMap{},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
			},
		},
		{
			name:              "case 2: empty current and non-empty desired, expected empty",
			currentConfigMaps: []*corev1.ConfigMap{},
			desiredConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-2",
						Namespace: metav1.NamespaceSystem,
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{},
		},
		{
			name: "case 3: equal current and desired, expected current",
			currentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-2",
						Namespace: metav1.NamespaceSystem,
					},
				},
			},
			desiredConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-2",
						Namespace: metav1.NamespaceSystem,
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-2",
						Namespace: metav1.NamespaceSystem,
					},
				},
			},
		},
		{
			name: "case 4: unequal current and desired, expected current",
			currentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-2",
						Namespace: metav1.NamespaceSystem,
					},
				},
			},
			desiredConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-2",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-3",
						Namespace: metav1.NamespaceSystem,
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-2",
						Namespace: metav1.NamespaceSystem,
					},
				},
			},
		},
		{
			name: "case 5: unequal current and desired in multiple namespaces, expected current",
			currentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-2",
						Namespace: metav1.NamespacePublic,
					},
				},
			},
			desiredConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespacePublic,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-2",
						Namespace: metav1.NamespacePublic,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-3",
						Namespace: metav1.NamespaceSystem,
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap-2",
						Namespace: metav1.NamespacePublic,
					},
				},
			},
		},
	}

	c := Config{
		Tenant:         &guestMock{},
		Logger:         microloggertest.New(),
		ProjectName:    "cluster-operator",
		RegistryDomain: "quay.io",
	}
	newService, err := New(c)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := newService.newDeleteChangeForDeletePatch(context.TODO(), tc.currentConfigMaps, tc.desiredConfigMaps)

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
