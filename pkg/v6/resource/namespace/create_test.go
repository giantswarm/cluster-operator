package namespace

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Resource_Namespace_newCreateChange(t *testing.T) {
	obj := v1alpha1.ClusterGuestConfig{
		DNSZone: "5xchu.aws.giantswarm.io",
		ID:      "5xchu",
		Owner:   "giantswarm",
	}

	testCases := []struct {
		name              string
		cur               interface{}
		des               interface{}
		expectedNamespace *apiv1.Namespace
	}{
		{
			name:              "case 0: nil current and desired, expected nil",
			cur:               nil,
			des:               nil,
			expectedNamespace: nil,
		},
		{
			name: "case 1: non-empty current, nil desired, expected nil",
			cur: &apiv1.Namespace{
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "giantswarm",
				},
			},
			des:               nil,
			expectedNamespace: nil,
		},
		{
			name: "case 2: nil current, non-empty desired, expected desired",
			cur:  nil,
			des: &apiv1.Namespace{
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "giantswarm",
				},
			},
			expectedNamespace: &apiv1.Namespace{
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "giantswarm",
				},
			},
		},
		{
			name: "case 3: equal current and desired, expected nil",
			cur: &apiv1.Namespace{
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "giantswarm",
				},
			},
			des: &apiv1.Namespace{
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "giantswarm",
				},
			},
			expectedNamespace: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				BaseClusterConfig: cluster.Config{
					ClusterID: "test-cluster",
				},
				Tenant:      &guestMock{},
				Logger:      microloggertest.New(),
				ProjectName: "cluster-operator",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
				ToClusterObjectMetaFunc: func(v interface{}) (apismetav1.ObjectMeta, error) {
					return v.(apismetav1.ObjectMeta), nil
				},
			}
			newResource, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := newResource.newCreateChange(context.TODO(), obj, tc.cur, tc.des)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if tc.expectedNamespace == nil {
				if tc.expectedNamespace != result {
					t.Fatalf("expected %#v, got %#v", tc.expectedNamespace, result)
				}
			} else {
				name := result.(*apiv1.Namespace).Name
				if tc.expectedNamespace.Name != name {
					t.Fatalf("expected %q, got %q", tc.expectedNamespace.Name, name)
				}
			}
		})
	}
}
