package chartconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/apimachinery/pkg/runtime"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
)

func Test_GetCurrentState(t *testing.T) {
	testCases := []struct {
		name                 string
		obj                  interface{}
		presentChartConfigs  []*v1alpha1.ChartConfig
		expectedChartConfigs []*v1alpha1.ChartConfig
	}{
		{
			name: "case 1: no results",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			presentChartConfigs:  []*v1alpha1.ChartConfig{},
			expectedChartConfigs: []*v1alpha1.ChartConfig{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tc.presentChartConfigs))
			for _, cc := range tc.presentChartConfigs {
				objs = append(objs, cc)
			}

			fakeGuestG8sClient := fake.NewSimpleClientset(objs...)
			guestClusterService := &guestClusterServiceMock{
				fakeGuestG8sClient: fakeGuestG8sClient,
			}

			c := Config{
				BaseClusterConfig: cluster.Config{
					ClusterID: "test-cluster",
				},
				G8sClient:           fake.NewSimpleClientset(),
				GuestClusterService: guestClusterService,
				K8sClient:           clientgofake.NewSimpleClientset(),
				Logger:              microloggertest.New(),
				ProjectName:         "cluster-operator",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
			}
			newResource, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := newResource.GetCurrentState(context.TODO(), tc.obj)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			chartConfigs, err := toChartConfigs(result)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if !reflect.DeepEqual(chartConfigs, tc.expectedChartConfigs) {
				t.Fatalf("expected %#v got %#v", tc.expectedChartConfigs, chartConfigs)
			}
		})
	}
}
