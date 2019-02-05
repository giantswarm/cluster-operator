package namespace

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
)

func Test_Resource_Namespace_GetDesiredState(t *testing.T) {
	testCases := []struct {
		name           string
		obj            interface{}
		expectedName   string
		expectedLabels map[string]string
	}{
		{
			name: "case 0: basic match",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			expectedName: "giantswarm",
			expectedLabels: map[string]string{
				"giantswarm.io/cluster":      "5xchu",
				"giantswarm.io/managed-by":   "cluster-operator",
				"giantswarm.io/organization": "giantswarm",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				BaseClusterConfig: cluster.Config{
					ClusterID: "test-cluster",
				},
				Logger:      microloggertest.New(),
				ProjectName: "cluster-operator",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
				Tenant: tenantclustertest.New(tenantclustertest.Config{}),
				ToClusterObjectMetaFunc: func(v interface{}) (apismetav1.ObjectMeta, error) {
					return v.(apismetav1.ObjectMeta), nil
				},
			}
			newResource, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := newResource.GetDesiredState(context.TODO(), tc.obj)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			name := result.(*apiv1.Namespace).Name
			if tc.expectedName != name {
				t.Fatalf("expected %q got %q", tc.expectedName, name)
			}

			labels := result.(*apiv1.Namespace).Labels
			if !reflect.DeepEqual(tc.expectedLabels, labels) {
				t.Fatalf("expected %#v got %#v", tc.expectedLabels, labels)
			}
		})
	}
}
