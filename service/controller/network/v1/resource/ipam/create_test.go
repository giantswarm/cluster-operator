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

func Test_EnsureCreated_common_cases(t *testing.T) {
	testCases := []struct {
		name              string
		clusterNetworkCfg interface{}
		network           net.IPNet
		reservedSubnets   []net.IPNet
		apiReactors       []k8stestingReactorFactoryFunc
		expectedStatus    *v1alpha1.ClusterNetworkConfigStatus
		errorMatcher      func(error) bool
	}{
		{
			name: "case 0: allocate first network",
			clusterNetworkCfg: &v1alpha1.ClusterNetworkConfig{
				Spec: v1alpha1.ClusterNetworkConfigSpec{
					Cluster: v1alpha1.ClusterNetworkConfigSpecCluster{
						Network: v1alpha1.ClusterNetworkConfigSpecClusterNetwork{
							MaskBits: 24,
						},
					},
				},
			},
			network:         mustParseNetworkCIDR("172.28.0.0/16"),
			reservedSubnets: []net.IPNet{},
			apiReactors: []k8stestingReactorFactoryFunc{
				verifyClusterNetworkConfigStatusUpdateReactor,
			},
			expectedStatus: &v1alpha1.ClusterNetworkConfigStatus{
				IP:   "172.28.0.0",
				Mask: "255.255.255.0",
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: allocate network after reservedSubnets",
			clusterNetworkCfg: &v1alpha1.ClusterNetworkConfig{
				Spec: v1alpha1.ClusterNetworkConfigSpec{
					Cluster: v1alpha1.ClusterNetworkConfigSpecCluster{
						Network: v1alpha1.ClusterNetworkConfigSpecClusterNetwork{
							MaskBits: 24,
						},
					},
				},
			},
			network:         mustParseNetworkCIDR("172.28.0.0/16"),
			reservedSubnets: []net.IPNet{mustParseNetworkCIDR("172.28.0.0/24")},
			apiReactors: []k8stestingReactorFactoryFunc{
				verifyClusterNetworkConfigStatusUpdateReactor,
			},
			expectedStatus: &v1alpha1.ClusterNetworkConfigStatus{
				IP:   "172.28.1.0",
				Mask: "255.255.255.0",
			},
			errorMatcher: nil,
		},
		{
			name: "case 2: return error when whole cluster network has been allocated",
			clusterNetworkCfg: &v1alpha1.ClusterNetworkConfig{
				Spec: v1alpha1.ClusterNetworkConfigSpec{
					Cluster: v1alpha1.ClusterNetworkConfigSpecCluster{
						Network: v1alpha1.ClusterNetworkConfigSpecClusterNetwork{
							MaskBits: 24,
						},
					},
				},
			},
			network:         mustParseNetworkCIDR("192.168.0.0/16"),
			reservedSubnets: []net.IPNet{mustParseNetworkCIDR("192.168.0.0/16")},
			apiReactors: []k8stestingReactorFactoryFunc{
				verifyClusterNetworkConfigStatusUpdateReactor,
			},
			expectedStatus: nil,
			errorMatcher:   ipam.IsSpaceExhausted,
		},
		{
			name: "case 3: don't allocate network when ClusterNetworkConfig.Status contains network IP",
			clusterNetworkCfg: &v1alpha1.ClusterNetworkConfig{
				Spec: v1alpha1.ClusterNetworkConfigSpec{
					Cluster: v1alpha1.ClusterNetworkConfigSpecCluster{
						Network: v1alpha1.ClusterNetworkConfigSpecClusterNetwork{
							MaskBits: 24,
						},
					},
				},
				Status: v1alpha1.ClusterNetworkConfigStatus{
					IP:   "192.168.2.0",
					Mask: "255.255.255.0",
				},
			},
			network:         mustParseNetworkCIDR("192.168.0.0/16"),
			reservedSubnets: []net.IPNet{mustParseNetworkCIDR("192.168.0.0/16")},
			apiReactors: []k8stestingReactorFactoryFunc{
				alwaysFailReactor,
			},
			expectedStatus: nil,
			errorMatcher:   nil,
		},
		{
			name: "case 4: handle unexpected k8s api error",
			clusterNetworkCfg: &v1alpha1.ClusterNetworkConfig{
				Spec: v1alpha1.ClusterNetworkConfigSpec{
					Cluster: v1alpha1.ClusterNetworkConfigSpecCluster{
						Network: v1alpha1.ClusterNetworkConfigSpecClusterNetwork{
							MaskBits: 24,
						},
					},
				},
			},
			network:         mustParseNetworkCIDR("192.168.0.0/16"),
			reservedSubnets: []net.IPNet{mustParseNetworkCIDR("192.168.0.0/24")},
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
				AllocatedSubnets: tc.reservedSubnets,
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

			err = r.EnsureCreated(context.TODO(), tc.clusterNetworkCfg)

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

func Test_EnsureCreated_releases_subnet_when_UpdateStatus_fails(t *testing.T) {
	testLogger := microloggertest.New()

	memoryStorage, err := memory.New(memory.DefaultConfig())
	if err != nil {
		t.Fatalf("memory storage creation failed: %v", err)
	}

	network := mustParseNetworkCIDR("172.28.0.0/16")

	ic := ipam.Config{
		Logger:           testLogger,
		Storage:          memoryStorage,
		Network:          &network,
		AllocatedSubnets: []net.IPNet{},
	}

	ipamSvc, err := ipam.New(ic)
	if err != nil {
		t.Fatalf("ipam service creation failed: %v", err)
	}

	reactors := []k8stesting.Reactor{alwaysReturnErrorReactor(unknownAPIError)(t, nil)}
	g8sClient := fake.NewSimpleClientset()
	g8sClient.ReactionChain = append(reactors, g8sClient.ReactionChain...)

	r := &Resource{
		g8sClient: g8sClient,
		ipam:      ipamSvc,
		logger:    testLogger,
	}

	clusterNetworkCfg := v1alpha1.ClusterNetworkConfig{
		Spec: v1alpha1.ClusterNetworkConfigSpec{
			Cluster: v1alpha1.ClusterNetworkConfigSpecCluster{
				Network: v1alpha1.ClusterNetworkConfigSpecClusterNetwork{
					MaskBits: 24,
				},
			},
		},
	}

	// EnsureCreated allocates subnet when Status.IP == "", but
	// UpdateStatus(clusterNetworkCfg) call returns error. Allocated subnet must
	// be released.
	err = r.EnsureCreated(context.TODO(), &clusterNetworkCfg)
	if microerror.Cause(err) != unknownAPIError {
		t.Fatalf("EnsureCreated() returned error %v, expected %v", err, unknownAPIError)
	}

	// When reallocating subnet, one freed above must be acquired.
	subnet, err := ipamSvc.CreateSubnet(context.TODO(), net.CIDRMask(24, 32), "")

	// NOTE: This expectation breaks when IPAM algorithm is changed to be
	// rotating next free instead of current first free based.
	expectedSubnetIP := "172.28.0.0"
	if subnet.IP.String() != expectedSubnetIP {
		t.Fatalf("reallocating subnet returned IP %s, expected %s", subnet.IP.String(), expectedSubnetIP)
	}

	expectedSubnetMask := "255.255.255.0"
	if net.IP(subnet.Mask).String() != expectedSubnetMask {
		t.Fatalf("reallocating subnet returned Mask %s, expected %s", net.IP(subnet.Mask).String(), expectedSubnetMask)
	}
}
