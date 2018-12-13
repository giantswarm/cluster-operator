package chart

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
)

func Test_Resource_Chart_newCreate(t *testing.T) {
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
			name:         "case 2: empty current, non-empty desired, expected desired",
			currentState: &ResourceState{},
			desiredState: &ResourceState{
				ChartName: "desired",
			},
			expectedState: &ResourceState{
				ChartName: "desired",
			},
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
			name: "case 4: different non-empty current and desired, expected empty",
			currentState: &ResourceState{
				ChartName: "current",
			},
			desiredState: &ResourceState{
				ChartName: "desired",
			},
			expectedState: &ResourceState{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helmClient := &helmMock{}

			c := Config{
				ApprClient: &apprMock{},
				BaseClusterConfig: cluster.Config{
					ClusterID: "test-cluster",
				},
				Fs:        afero.NewMemMapFs(),
				G8sClient: fake.NewSimpleClientset(),
				Tenant: &guestMock{
					fakeGuestHelmClient: helmClient,
				},
				K8sClient:      clientgofake.NewSimpleClientset(),
				Logger:         microloggertest.New(),
				ProjectName:    "cluster-operator",
				RegistryDomain: "quay.io",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
			}

			newResource, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := newResource.newCreateChange(context.TODO(), obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			createChange, ok := result.(*ResourceState)
			if !ok {
				t.Fatalf("expected '%T', got '%T'", createChange, result)
			}
			if !reflect.DeepEqual(createChange, tc.expectedState) {
				t.Fatalf("ChartState == %q, want %q", createChange, tc.expectedState)
			}
		})
	}
}
