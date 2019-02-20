package chartoperator

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/apprclient/apprclienttest"
	"github.com/giantswarm/helmclient/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
)

func Test_Resource_Chart_newUpdate(t *testing.T) {
	obj := v1alpha1.ClusterGuestConfig{
		DNSZone: "5xchu.aws.giantswarm.io",
		ID:      "5xchu",
		Owner:   "giantswarm",
	}

	testCases := []struct {
		currentState  *ResourceState
		desiredState  *ResourceState
		expectedState *ResourceState
		name          string
	}{
		{
			name:          "case 0: empty current and desired, expected empty",
			currentState:  &ResourceState{},
			desiredState:  &ResourceState{},
			expectedState: &ResourceState{},
		},
		{
			name: "case 1: non-empty current, empty desired, expected empty",
			currentState: &ResourceState{
				ChartName: "current",
			},
			desiredState:  &ResourceState{},
			expectedState: &ResourceState{},
		},

		{
			name:         "case 2: empty current, non-empty desired, expected empty",
			currentState: &ResourceState{},
			desiredState: &ResourceState{
				ChartName: "desired",
			},
			expectedState: &ResourceState{},
		},
		{
			name: "case 3: equal non-empty current and desired, expected empty",
			currentState: &ResourceState{
				ChartName: "desired",
			},
			desiredState: &ResourceState{
				ChartName: "desired",
			},
			expectedState: &ResourceState{},
		},
		{
			name: "case 4: different non-empty current and desired, expected desired",
			currentState: &ResourceState{
				ChartName:      "current",
				ReleaseVersion: "0.1.2",
			},
			desiredState: &ResourceState{
				ChartName:      "desired",
				ReleaseVersion: "0.1.3",
			},
			expectedState: &ResourceState{
				ChartName:      "desired",
				ReleaseVersion: "0.1.3",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helmClient := helmclienttest.New(helmclienttest.Config{})

			c := Config{
				ApprClient: apprclienttest.New(apprclienttest.Config{}),
				BaseClusterConfig: cluster.Config{
					ClusterID: "test-cluster",
				},
				ClusterIPRange: "172.31.0.0/16",
				Fs:             afero.NewMemMapFs(),
				G8sClient:      fake.NewSimpleClientset(),
				K8sClient:      clientgofake.NewSimpleClientset(),
				Logger:         microloggertest.New(),
				ProjectName:    "cluster-operator",
				RegistryDomain: "quay.io",
				Tenant: &tenantMock{
					fakeTenantHelmClient: helmClient,
				},
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
				ToClusterObjectMetaFunc: func(v interface{}) (metav1.ObjectMeta, error) {
					return metav1.ObjectMeta{}, nil
				},
			}

			newResource, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := newResource.newUpdateChange(context.TODO(), obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			updateChange, ok := result.(*ResourceState)
			if !ok {
				t.Fatalf("expected '%T', got '%T'", updateChange, result)
			}
			if !reflect.DeepEqual(updateChange, tc.expectedState) {
				t.Fatalf("ChartState == %q, want %q", updateChange, tc.expectedState)
			}
		})
	}
}
