package configmap

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_ConfigMap_GetCurrentState(t *testing.T) {
	testCases := []struct {
		name               string
		config             ConfigMapConfig
		presentConfigMaps  []*corev1.ConfigMap
		expectedConfigMaps []*corev1.ConfigMap
	}{
		{
			name: "case 0: no results",
			config: ConfigMapConfig{
				ClusterID:      "5xchu",
				GuestAPIDomain: "5xchu.aws.giantswarm.io",
				Namespaces:     []string{},
			},
			presentConfigMaps:  []*corev1.ConfigMap{},
			expectedConfigMaps: []*corev1.ConfigMap{},
		},
		{
			name: "case 1: single result",
			config: ConfigMapConfig{
				ClusterID:      "5xchu",
				GuestAPIDomain: "5xchu.aws.giantswarm.io",
				Namespaces:     []string{},
			},
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
		},
		{
			name: "case 2: multiple results",
			config: ConfigMapConfig{
				ClusterID:      "5xchu",
				GuestAPIDomain: "5xchu.aws.giantswarm.io",
				Namespaces:     []string{},
			},
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "another-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "another-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
		},
		{
			name: "case 3: multiple namespaces, single result",
			config: ConfigMapConfig{
				ClusterID:      "5xchu",
				GuestAPIDomain: "5xchu.aws.giantswarm.io",
				Namespaces: []string{
					"giantswarm",
				},
			},
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
		},
		{
			name: "case 4: multiple namespaces, multiple results",
			config: ConfigMapConfig{
				ClusterID:      "5xchu",
				GuestAPIDomain: "5xchu.aws.giantswarm.io",
				Namespaces: []string{
					"giantswarm",
				},
			},
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tc.presentConfigMaps))
			for _, cc := range tc.presentConfigMaps {
				objs = append(objs, cc)
			}

			fakeGuestK8sClient := fake.NewSimpleClientset(objs...)
			guestService := &guestMock{
				fakeGuestK8sClient: fakeGuestK8sClient,
			}

			c := Config{
				Guest:          guestService,
				Logger:         microloggertest.New(),
				ProjectName:    "cluster-operator",
				RegistryDomain: "quay.io",
			}
			newService, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			configMaps, err := newService.GetCurrentState(context.TODO(), tc.config)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if len(configMaps) != len(tc.expectedConfigMaps) {
				t.Fatalf("expected %d configsmaps got %d", len(tc.expectedConfigMaps), len(configMaps))
			}

			for _, cm := range configMaps {
				found := false
				for _, ec := range tc.expectedConfigMaps {
					if reflect.DeepEqual(cm, ec) {
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("unexpected configmap %#v among returned values", *cm)
				}
			}

		})
	}
}
