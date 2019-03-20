package kubeconfig

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	clientgofake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"github.com/giantswarm/cluster-operator/pkg/v13/chartconfig"
)

func Test_Resource_GetDesiredState(t *testing.T) {
	tests := []struct {
		name           string
		config         chartconfig.ClusterConfig
		expectedSecret *corev1.Secret
		errorMatcher   func(error) bool
		secretCert     *corev1.Secret
	}{
		{
			name: "case 0: cluster config",
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
			config: chartconfig.ClusterConfig{
				APIDomain:    "http://www.giantswarm.io",
				ClusterID:    "w7utg",
				Organization: "giantswarm",
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "w7utg-kubeconfig",
					Namespace: "giantswarm",
					Labels: map[string]string{
						"giantswarm.io/managed-by": "cluster-operator",
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
					fmt.Println(string(secret.Data["kubeConfig"]))
					fmt.Println("fuck")
					t.Fatalf("want matching data \n %s", cmp.Diff(secret.Data, tc.expectedSecret.Data))
				}
				if !reflect.DeepEqual(secret.TypeMeta, tc.expectedSecret.TypeMeta) {
					t.Fatalf("want matching typemeta \n %s", cmp.Diff(secret.TypeMeta, tc.expectedSecret.TypeMeta))
				}
			}
		})
	}
}
