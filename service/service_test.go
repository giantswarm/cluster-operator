package service

import (
	"net"
	"reflect"
	"testing"
)

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
