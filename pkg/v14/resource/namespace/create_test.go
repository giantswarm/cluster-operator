package namespace

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/tenantcluster/tenantclustertest"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
)

func Test_Resource_Namespace_desiredNamespace(t *testing.T) {
	testCases := []struct {
		name              string
		obj               v1alpha1.ClusterGuestConfig
		expectedNamespace *corev1.Namespace
	}{
		{
			name: "case 0: basic match",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			expectedNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm",
					Labels: map[string]string{
						"giantswarm.io/cluster":      "5xchu",
						"giantswarm.io/managed-by":   "cluster-operator",
						"giantswarm.io/organization": "giantswarm",
					},
				},
			},
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				BaseClusterConfig: cluster.Config{
					ClusterID: "test-cluster",
				},
				Logger:      microloggertest.New(),
				ProjectName: "cluster-operator",
				Tenant:      tenantclustertest.New(tenantclustertest.Config{}),
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
				ToClusterObjectMetaFunc: func(v interface{}) (metav1.ObjectMeta, error) {
					return v.(metav1.ObjectMeta), nil
				},
			}
			newResource, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := newResource.desiredNamespace(ctx, tc.obj)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if !reflect.DeepEqual(result.ObjectMeta, tc.expectedNamespace.ObjectMeta) {
				t.Fatalf("want matching objectmeta \n %s", cmp.Diff(result.ObjectMeta, tc.expectedNamespace.ObjectMeta))
			}

			if !reflect.DeepEqual(result.ObjectMeta, tc.expectedNamespace.ObjectMeta) {
				t.Fatalf("want matching objectmeta \n %s", cmp.Diff(result.ObjectMeta, tc.expectedNamespace.ObjectMeta))
			}
		})
	}
}
