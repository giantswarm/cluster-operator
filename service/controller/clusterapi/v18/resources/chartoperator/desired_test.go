package chartoperator

import (
	"context"
	"reflect"
	"testing"

	g8sfake "github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/apprclient/apprclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
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
			obj:  v1alpha1.Cluster{},
			expectedState: ResourceState{
				ChartName: "chart-operator-chart",
				ChartValues: Values{
					ClusterDNSIP: "172.31.0.10",
					Image: Image{
						Registry: "quay.io",
					},
					Tiller: Tiller{
						Namespace: "giantswarm",
					},
				},
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
				Fs:         afero.NewMemMapFs(),
				G8sClient:  g8sfake.NewSimpleClientset(),
				K8sClient:  k8sfake.NewSimpleClientset(),
				Logger:     microloggertest.New(),

				DNSIP:          "172.31.0.10",
				RegistryDomain: "quay.io",
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
				t.Fatalf("want matching ResourceState \n %s", cmp.Diff(chartState, tc.expectedState))
			}
		})
	}
}
