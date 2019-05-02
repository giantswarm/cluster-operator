package clusterconfigmap

import (
	"context"
	"github.com/giantswarm/microerror"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_GetDesiredState(t *testing.T) {
	tests := []struct {
		name              string
		config            *v1alpha1.AWSClusterConfig
		expectedConfigMap *corev1.ConfigMap
		errorMatcher      func(error) bool
	}{
		{
			name: "case 0: flawless flow",
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
			expectedConfigMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "w7utg-cluster-values",
					Namespace: "w7utg",
					Labels: map[string]string{
						"giantswarm.io/cluster":      "w7utg",
						"giantswarm.io/organization": "giantswarm",
						"giantswarm.io/service-type": "managed",
						"giantswarm.io/managed-by":   "cluster-operator",
					},
				},
				Data: map[string]string{
					"values": "baseDomain: giantswarm.io\n",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				GetClusterConfigFunc: toClusterConfigCR,
				K8sClient:            k8sfake.NewSimpleClientset(),
				Logger:               microloggertest.New(),

				ProjectName: "cluster-operator",
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
				configMap, err := toConfigMap(result[0])
				if err != nil {
					t.Fatalf("error == %#v, want nil", err)
				}

				if !reflect.DeepEqual(configMap.ObjectMeta, tc.expectedConfigMap.ObjectMeta) {
					t.Fatalf("want matching objectmeta \n %s", cmp.Diff(configMap.ObjectMeta, tc.expectedConfigMap.ObjectMeta))
				}
				if !reflect.DeepEqual(configMap.Data, tc.expectedConfigMap.Data) {
					t.Fatalf("want matching data \n %s", cmp.Diff(configMap.Data, tc.expectedConfigMap.Data))
				}
				if !reflect.DeepEqual(configMap.TypeMeta, tc.expectedConfigMap.TypeMeta) {
					t.Fatalf("want matching typemeta \n %s", cmp.Diff(configMap.TypeMeta, tc.expectedConfigMap.TypeMeta))
				}
			}
		})
	}
}

func toClusterConfigCR(obj interface{}) (v1alpha1.ClusterGuestConfig, error) {
	customConfig, ok := obj.(*v1alpha1.AWSClusterConfig)
	if !ok {
		return v1alpha1.ClusterGuestConfig{}, microerror.Mask(wrongTypeError)
	}
	return customConfig.Spec.Guest.ClusterGuestConfig, nil
}
