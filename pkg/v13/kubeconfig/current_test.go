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
		obj            interface{}
		expectedSecret *corev1.Secret
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: aws cluster config",
			obj: &v1alpha1.AWSClusterConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm-aws-tenant",
				},
				Spec: v1alpha1.AWSClusterConfigSpec{
					Guest: v1alpha1.AWSClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							DNSZone: "http://www.giantswarm.io",
							ID:      "w7utg",
							Owner:   "giantswarm",
						},
					},
				},
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "giantswarm-tenant",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"kubeConfig": []byte(kubeconfigYaml),
				},
			},
		},
		{
			name: "case 1: azure cluster config",
			obj: &v1alpha1.AzureClusterConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm-aws-tenant",
				},
				Spec: v1alpha1.AzureClusterConfigSpec{
					Guest: v1alpha1.AzureClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							DNSZone: "http://www.giantswarm.io",
							ID:      "w7utg",
							Owner:   "giantswarm",
						},
					},
				},
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "giantswarm-tenant",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"kubeConfig": []byte(kubeconfigYaml),
				},
			},
		},
		{
			name: "case 2: kvm cluster config",
			obj: &v1alpha1.KVMClusterConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm-aws-tenant",
				},
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							DNSZone: "http://www.giantswarm.io",
							ID:      "w7utg",
							Owner:   "giantswarm",
						},
					},
				},
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "giantswarm-tenant",
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
				ResourceName:      "giantswarm-tenant",
				ResourceNamespace: "giantswarm",
			}
			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := r.GetCurrentState(context.Background(), tc.obj)
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
