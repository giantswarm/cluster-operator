package namespace

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
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
		expectedNamespace *corev1.Namespace
	}{
		{
			name:              "case 0: nil current and desired, expected nil",
			cur:               nil,
			des:               nil,
			expectedNamespace: nil,
		},
		{
			name: "case 1: non-empty current, nil desired, expected nil",
			cur: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm",
				},
			},
			des:               nil,
			expectedNamespace: nil,
		},
		{
			name: "case 2: nil current, non-empty desired, expected desired",
			cur:  nil,
			des: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm",
				},
			},
			expectedNamespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm",
				},
			},
		},
		{
			name: "case 3: equal current and desired, expected nil",
			cur: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "giantswarm",
				},
			},
			des: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
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
				Logger: microloggertest.New(),
				Tenant: &tenantMock{},
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

			result, err := newResource.newCreateChange(context.TODO(), obj, tc.cur, tc.des)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if tc.expectedNamespace == nil {
				if tc.expectedNamespace != result {
					t.Fatalf("expected %#v, got %#v", tc.expectedNamespace, result)
				}
			} else {
				name := result.(*corev1.Namespace).Name
				if tc.expectedNamespace.Name != name {
					t.Fatalf("expected %q, got %q", tc.expectedNamespace.Name, name)
				}
			}
		})
	}
}
