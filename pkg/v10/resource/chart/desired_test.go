package chart

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/apprclient/apprclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/tenantcluster/tenantclustertest"
	"github.com/spf13/afero"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
)

func Test_Chart_GetDesiredState(t *testing.T) {
	testCases := []struct {
		name          string
		obj           interface{}
		expectedState ResourceState
		errorMatcher  func(error) bool
	}{
		{
			name: "case 0: basic match",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			expectedState: ResourceState{
				ChartName:      "chart-operator-chart",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.2",
				ReleaseStatus:  "DEPLOYED",
			},
		},
	}

	var apprClient apprclient.Interface
	{
		c := apprclienttest.Config{
			DefaultReleaseVersion: "0.1.2",
		}
		apprClient = apprclienttest.New(c)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				ApprClient: apprClient,
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
				Tenant:         tenantclustertest.New(tenantclustertest.Config{}),
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
			}

			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := r.GetDesiredState(context.TODO(), tc.obj)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			chartState, err := toResourceState(result)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			if !reflect.DeepEqual(chartState, tc.expectedState) {
				t.Fatalf("ChartState == %q, want %q", chartState, tc.expectedState)
			}
		})
	}
}
