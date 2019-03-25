package service

import (
	"net"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/viper"

	"github.com/giantswarm/cluster-operator/flag"
)

func Test_Service_New(t *testing.T) {
	testCases := []struct {
		description          string
		config               func() Config
		expectedErrorHandler func(error) bool
	}{
		{
			description: "empty value config must return invalidConfigError",
			config: func() Config {
				return Config{}
			},
			expectedErrorHandler: IsInvalidConfig,
		},
		{
			description: "production-like config must be valid",
			config: func() Config {
				config := Config{}

				config.Logger = microloggertest.New()

				config.Flag = flag.New()
				config.Viper = viper.New()

				config.Description = "test"
				config.GitCommit = "test"
				config.ProjectName = "test"
				config.Source = "test"

				config.Viper.Set(config.Flag.Guest.Cluster.Calico.CIDR, "16")
				config.Viper.Set(config.Flag.Guest.Cluster.Calico.Subnet, "172.26.0.0")
				config.Viper.Set(config.Flag.Guest.Cluster.Kubernetes.API.ClusterIPRange, "172.31.0.0/16")
				config.Viper.Set(config.Flag.Service.Image.Registry.Domain, "quay.io")
				config.Viper.Set(config.Flag.Service.KubeConfig.Resource.Namespace, "giantswarm")
				config.Viper.Set(config.Flag.Service.Kubernetes.Address, "http://127.0.0.1:6443")
				config.Viper.Set(config.Flag.Service.Kubernetes.InCluster, "false")

				return config
			},
			expectedErrorHandler: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := New(tc.config())

			if err != nil {
				if tc.expectedErrorHandler == nil {
					t.Fatalf("unexpected error returned: %#v", err)
				}
				if !tc.expectedErrorHandler(err) {
					t.Fatalf("incorrect error returned: %#v", err)
				}
			}
		})
	}
}

func Test_parseClusterIPRange(t *testing.T) {
	testCases := []struct {
		name                string
		inputCIDR           string
		expectedNetworkIP   net.IP
		expectedAPIServerIP net.IP
		errorMatcher        func(error) bool
	}{
		{
			name:                "case 0: valid /16 network",
			inputCIDR:           "172.31.0.0/16",
			expectedNetworkIP:   net.IPv4(172, 31, 0, 0),
			expectedAPIServerIP: net.IPv4(172, 31, 0, 1),
			errorMatcher:        nil,
		},
		{
			name:                "case 1: valid /24 network",
			inputCIDR:           "192.168.12.0/24",
			expectedNetworkIP:   net.IPv4(192, 168, 12, 0),
			expectedAPIServerIP: net.IPv4(192, 168, 12, 1),
			errorMatcher:        nil,
		},
		{
			name:                "case 2: valid /24 network",
			inputCIDR:           "192.168.12.16/24",
			expectedNetworkIP:   net.IPv4(192, 168, 12, 0),
			expectedAPIServerIP: net.IPv4(192, 168, 12, 1),
			errorMatcher:        nil,
		},
		{
			name:                "case 3: invalid /25 network",
			inputCIDR:           "172.31.0.0/25",
			expectedNetworkIP:   nil,
			expectedAPIServerIP: nil,
			errorMatcher:        IsInvalidConfig,
		},
		{
			name:                "case 4: invalid /27 network",
			inputCIDR:           "172.31.0.0/27",
			expectedNetworkIP:   nil,
			expectedAPIServerIP: nil,
			errorMatcher:        IsInvalidConfig,
		},
		{
			name:                "case 5: invalid IPv6 network",
			inputCIDR:           "2001:db8:a0b:12f0::1/32",
			expectedNetworkIP:   nil,
			expectedAPIServerIP: nil,
			errorMatcher:        IsInvalidConfig,
		},
		{
			name:                "case 6: invalid IPv4 network mask",
			inputCIDR:           "172.0.0.1/33",
			expectedNetworkIP:   nil,
			expectedAPIServerIP: nil,
			errorMatcher:        IsInvalidConfig,
		},
		{
			name:                "case 6: invalid CIDR",
			inputCIDR:           "256.0.0.1/33",
			expectedNetworkIP:   nil,
			expectedAPIServerIP: nil,
			errorMatcher:        IsInvalidConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			networkIP, apiServerIP, err := parseClusterIPRange(tc.inputCIDR)

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

			// Force IPs to same representation for comparison.
			networkIP = networkIP.To4()
			tc.expectedNetworkIP = tc.expectedNetworkIP.To4()
			apiServerIP = apiServerIP.To4()
			tc.expectedAPIServerIP = tc.expectedAPIServerIP.To4()

			if !reflect.DeepEqual(networkIP, tc.expectedNetworkIP) ||
				!reflect.DeepEqual(apiServerIP, tc.expectedAPIServerIP) {
				t.Fatalf("NetworkIP == %q, want %q, APIServerIP == %q, want %q",
					networkIP, tc.expectedNetworkIP, apiServerIP, tc.expectedAPIServerIP)
			}
		})
	}
}
