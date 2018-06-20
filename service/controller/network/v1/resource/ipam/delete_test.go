package ipam

import (
	"context"
	"net"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/microstorage/memory"
	k8stesting "k8s.io/client-go/testing"
)

func Test_EnsureDeleted_common_cases(t *testing.T) {
	testCases := []struct {
		name              string
		clusterNetworkCfg interface{}
		network           net.IPNet
		allocatedSubnets  []net.IPNet
		apiReactors       []k8stestingReactorFactoryFunc
		expectedStatus    *v1alpha1.ClusterNetworkConfigStatus
		errorMatcher      func(error) bool
	}{
		{
			name: "case 0: free allocated subnet",
			clusterNetworkCfg: &v1alpha1.ClusterNetworkConfig{
				Spec: v1alpha1.ClusterNetworkConfigSpec{
					Cluster: v1alpha1.ClusterNetworkConfigSpecCluster{
						Network: v1alpha1.ClusterNetworkConfigSpecClusterNetwork{
							MaskBits: 24,
						},
					},
				},
				Status: v1alpha1.ClusterNetworkConfigStatus{
					IP:   "172.28.0.0",
					Mask: "255.255.255.0",
				},
			},
			network:          mustParseNetworkCIDR("172.28.0.0/16"),
			allocatedSubnets: []net.IPNet{mustParseNetworkCIDR("172.28.0.0/24")},
			apiReactors: []k8stestingReactorFactoryFunc{
				verifyClusterNetworkConfigStatusUpdateReactor,
			},
			expectedStatus: &v1alpha1.ClusterNetworkConfigStatus{
				IP:   "",
				Mask: "",
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: free subnet that hasn't been allocated",
			clusterNetworkCfg: &v1alpha1.ClusterNetworkConfig{
				Spec: v1alpha1.ClusterNetworkConfigSpec{
					Cluster: v1alpha1.ClusterNetworkConfigSpecCluster{
						Network: v1alpha1.ClusterNetworkConfigSpecClusterNetwork{
							MaskBits: 24,
						},
					},
				},
				Status: v1alpha1.ClusterNetworkConfigStatus{
					IP:   "172.28.0.0",
					Mask: "255.255.255.0",
				},
			},
			network:          mustParseNetworkCIDR("172.28.0.0/16"),
			allocatedSubnets: []net.IPNet{},
			apiReactors: []k8stestingReactorFactoryFunc{
				verifyClusterNetworkConfigStatusUpdateReactor,
			},
			expectedStatus: &v1alpha1.ClusterNetworkConfigStatus{
				IP:   "",
				Mask: "",
			},
			errorMatcher: nil,
		},
		{
			name: "case 2: handle unexpected k8s api error",
			clusterNetworkCfg: &v1alpha1.ClusterNetworkConfig{
				Spec: v1alpha1.ClusterNetworkConfigSpec{
					Cluster: v1alpha1.ClusterNetworkConfigSpecCluster{
						Network: v1alpha1.ClusterNetworkConfigSpecClusterNetwork{
							MaskBits: 24,
						},
					},
				},
				Status: v1alpha1.ClusterNetworkConfigStatus{
					IP:   "172.28.0.0",
					Mask: "255.255.255.0",
				},
			},
			network:          mustParseNetworkCIDR("192.168.0.0/16"),
			allocatedSubnets: []net.IPNet{mustParseNetworkCIDR("192.168.0.0/24")},
			apiReactors: []k8stestingReactorFactoryFunc{
				alwaysReturnErrorReactor(unknownAPIError),
			},
			expectedStatus: nil,
			errorMatcher:   func(err error) bool { return microerror.Cause(err) == unknownAPIError },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testLogger := microloggertest.New()

			memoryStorage, err := memory.New(memory.DefaultConfig())
			if err != nil {
				t.Fatalf("memory storage creation failed: %v", err)
			}

			ic := ipam.Config{
				Logger:           testLogger,
				Storage:          memoryStorage,
				Network:          &tc.network,
				AllocatedSubnets: tc.allocatedSubnets,
			}

			ipamSvc, err := ipam.New(ic)
			if err != nil {
				t.Fatalf("ipam service creation failed: %v", err)
			}

			var reactors []k8stesting.Reactor
			for _, f := range tc.apiReactors {
				reactors = append(reactors, f(t, tc.expectedStatus))
			}

			g8sClient := fake.NewSimpleClientset()
			g8sClient.ReactionChain = append(reactors, g8sClient.ReactionChain...)

			r := &Resource{
				g8sClient: g8sClient,
				ipam:      ipamSvc,
				logger:    testLogger,
			}

			err = r.EnsureDeleted(context.TODO(), tc.clusterNetworkCfg)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

		})
	}
}
