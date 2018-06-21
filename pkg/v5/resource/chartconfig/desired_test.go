package chartconfig

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
)

func Test_ChartConfig_GetDesiredState(t *testing.T) {
	testCases := []struct {
		name                     string
		obj                      interface{}
		provider                 string
		expectedChartConfigNames []string
	}{
		{
			name: "basic match for aws",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.aws.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			provider: label.ProviderAWS,
			expectedChartConfigNames: []string{
				"kubernetes-node-exporter-chart",
				"kubernetes-kube-state-metrics-chart",
			},
		},
		{
			name: "basic match for kvm",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.kvm.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			provider: label.ProviderKVM,
			expectedChartConfigNames: []string{
				"kubernetes-node-exporter-chart",
				"kubernetes-kube-state-metrics-chart",
			},
		},
		{
			name: "azure also has provider specific chartconfigs",
			obj: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.azure.giantswarm.io",
				ID:      "5xchu",
				Owner:   "giantswarm",
			},
			provider: label.ProviderAzure,
			expectedChartConfigNames: []string{
				"kubernetes-node-exporter-chart",
				"kubernetes-kube-state-metrics-chart",
				"kubernetes-nginx-ingress-controller-chart",
				"kubernetes-external-dns-chart",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				BaseClusterConfig: cluster.Config{
					ClusterID: "test-cluster",
				},
				G8sClient:   fake.NewSimpleClientset(),
				Guest:       &guestMock{},
				K8sClient:   clientgofake.NewSimpleClientset(),
				Logger:      microloggertest.New(),
				ProjectName: "cluster-operator",
				Provider:    tc.provider,
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
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

			chartConfigs, err := toChartConfigs(result)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if len(chartConfigs) != len(tc.expectedChartConfigNames) {
				t.Fatal("expected", len(tc.expectedChartConfigNames), "got", len(chartConfigs))
			}

			for _, expectedName := range tc.expectedChartConfigNames {
				_, err := getChartConfigByName(chartConfigs, expectedName)
				if IsNotFound(err) {
					t.Fatalf("expected chart %s not found", expectedName)
				} else if err != nil {
					t.Fatalf("expected nil, got %#v", err)
				}
			}
		})
	}
}
