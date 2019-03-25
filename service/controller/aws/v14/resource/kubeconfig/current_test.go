package kubeconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgofake "k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_GetCurrentState(t *testing.T) {
	tests := []struct {
		name           string
		config         *v1alpha1.AWSClusterConfig
		expectedSecret *corev1.Secret
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: aws cluster config",
			config: &v1alpha1.AWSClusterConfig{
				Spec: v1alpha1.AWSClusterConfigSpec{
					Guest: v1alpha1.AWSClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							DNSZone: "giantswarm.io",
							ID:      "w7utg",
							Name:    "My own snowflake cluster",
							Owner:   "giantswarm",
						},
					},
				},
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "w7utg-kubeconfig",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"kubeConfig": []byte(kubeconfigYaml),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			objs := make([]runtime.Object, 0, 0)
			if tc.expectedSecret != nil {
				objs = append(objs, tc.expectedSecret)
			}

			client := clientgofake.NewSimpleClientset(objs...)

			c := Config{
				K8sClient: client,
				Logger:    microloggertest.New(),

				ProjectName:       "cluster-operator",
				ResourceNamespace: "giantswarm",
			}
			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := r.GetCurrentState(context.Background(), tc.config)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if err == nil && tc.errorMatcher == nil {
				secret, err := toSecret(result[0])
				if err != nil {
					t.Fatalf("error == %#v, want nil", err)
				}

				if !reflect.DeepEqual(secret.ObjectMeta, tc.expectedSecret.ObjectMeta) {
					t.Fatalf("want matching objectmeta \n %s", cmp.Diff(secret.ObjectMeta, tc.expectedSecret.ObjectMeta))
				}
				if !reflect.DeepEqual(secret.Data, tc.expectedSecret.Data) {
					t.Fatalf("want matching data \n %s", cmp.Diff(secret.Data, tc.expectedSecret.Data))
				}
				if !reflect.DeepEqual(secret.TypeMeta, tc.expectedSecret.TypeMeta) {
					t.Fatalf("want matching typemeta \n %s", cmp.Diff(secret.TypeMeta, tc.expectedSecret.TypeMeta))
				}
			}
		})
	}
}
