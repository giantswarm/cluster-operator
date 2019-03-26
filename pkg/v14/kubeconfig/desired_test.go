package kubeconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgofake "k8s.io/client-go/kubernetes/fake"
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
		name            string
		certs           certs.AppOperator
		certsError      error
		config          *v1alpha1.AWSClusterConfig
		expectedSecrets []*corev1.Secret
		errorMatcher    func(error) bool
	}{
		{
			name: "case 0: basic match",
			certs: certs.AppOperator{
				APIServer: certs.TLS{
					CA:  []byte("ca"),
					Crt: []byte("crt"),
					Key: []byte("key"),
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
			expectedSecrets: []*corev1.Secret{
				{
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
		},
		{
			name:       "case 1: cert timeout, reconciliation stop",
			certsError: timeoutError,
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
			errorMatcher: IsTimeout,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			var ct certs.Interface
			{
				c := certstest.Config{
					AppOperator:      tc.certs,
					AppOperatorError: tc.certsError,
				}
				ct = certstest.NewSearcher(c)
			}

			c := Config{
				CertSearcher:         ct,
				K8sClient:            clientgofake.NewSimpleClientset(),
				Logger:               microloggertest.New(),
				GetClusterConfigFunc: toCR,
				ProjectName:          "cluster-operator",
				ResourceNamespace:    "giantswarm",
			}

			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

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
				if len(result) == 1 {
					secret, err := toSecret(result[0])
					if err != nil {
						t.Fatalf("error == %#v, want nil", err)
					}

					if !reflect.DeepEqual(secret.ObjectMeta, tc.expectedSecrets[0].ObjectMeta) {
						t.Fatalf("want matching objectmeta \n %s", cmp.Diff(secret.ObjectMeta, tc.expectedSecrets[0].ObjectMeta))
					}
					if !reflect.DeepEqual(secret.Data, tc.expectedSecrets[0].Data) {
						t.Fatalf("want matching data \n %s", cmp.Diff(secret.Data, tc.expectedSecrets[0].Data))
					}
					if !reflect.DeepEqual(secret.TypeMeta, tc.expectedSecrets[0].TypeMeta) {
						t.Fatalf("want matching typemeta \n %s", cmp.Diff(secret.TypeMeta, tc.expectedSecrets[0].TypeMeta))
					}
				}
				if len(tc.expectedSecrets) != len(result) {
					t.Fatalf("unmatch length between expected (%d) and result (%d)", len(tc.expectedSecrets), len(result))
				}
			}
		})
	}
}

func toCR(obj interface{}) (v1alpha1.ClusterGuestConfig, error) {
	customConfig, ok := obj.(*v1alpha1.AWSClusterConfig)
	if !ok {
		return v1alpha1.ClusterGuestConfig{}, microerror.Mask(wrongTypeError)
	}
	return customConfig.Spec.Guest.ClusterGuestConfig, nil
}
