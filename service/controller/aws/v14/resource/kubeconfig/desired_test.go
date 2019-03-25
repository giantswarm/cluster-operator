package kubeconfig

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	clientgofake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

const (
	kubeconfigYaml = `apiVersion: v1
kind: Config
clusters:
- name: giantswarm-w7utg
  cluster:
    server: api.giantswarm.io
    certificate-authority-data: Y2E=
users:
- name: giantswarm-w7utg-user
  user:
    client-certificate-data: Y3J0
    client-key-data: a2V5
contexts:
- name: giantswarm-w7utg-context
  context:
    cluster: giantswarm-w7utg
    user: giantswarm-w7utg-user
current-context: giantswarm-w7utg-context
preferences: {}
`
)

func Test_Resource_GetDesiredState(t *testing.T) {
	tests := []struct {
		name           string
		config         *v1alpha1.AWSClusterConfig
		expectedSecret *corev1.Secret
		errorMatcher   func(error) bool
		secretCert     *corev1.Secret
	}{
		{
			name: "case 0: basic match",
			secretCert: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterComponent": string(certs.AppOperatorAPICert),
						"clusterID":        "w7utg",
					},
				},
				Data: map[string][]byte{
					"ca":  []byte("ca"),
					"crt": []byte("crt"),
					"key": []byte("key"),
				},
			},
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
					Labels: map[string]string{
						"giantswarm.io/cluster":      "w7utg",
						"giantswarm.io/organization": "giantswarm",
						"giantswarm.io/managed-by":   "cluster-operator",
						"giantswarm.io/service-type": "managed",
					},
				},
				Data: map[string][]byte{
					"kubeConfig": []byte(kubeconfigYaml),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			client := clientgofake.NewSimpleClientset()
			fakeWatch := watch.NewFake()
			client.PrependWatchReactor("secrets", ktesting.DefaultWatchReactor(fakeWatch, nil))

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

			go func() {
				time.Sleep(2 * time.Second)
				fakeWatch.Add(tc.secretCert)
			}()
			result, err := r.GetDesiredState(context.Background(), tc.config)
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
