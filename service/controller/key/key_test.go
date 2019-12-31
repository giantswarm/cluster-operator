package key

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_CertConfigCertOperatorVersion(t *testing.T) {
	testCases := []struct {
		description     string
		certConfig      v1alpha1.CertConfig
		expectedVersion string
	}{
		{
			description:     "empty value",
			certConfig:      v1alpha1.CertConfig{},
			expectedVersion: "",
		},
		{
			description: "CertConfig with version",
			certConfig: v1alpha1.CertConfig{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"cert-operator.giantswarm.io/version": "1.0.1",
					},
				},
			},
			expectedVersion: "1.0.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			version := CertConfigCertOperatorVersion(tc.certConfig)
			if version != tc.expectedVersion {
				t.Fatalf("version '%s' doesn't match expected '%s'", version, tc.expectedVersion)
			}
		})
	}
}

func Test_DNSIP(t *testing.T) {
	testCases := []struct {
		description  string
		input        string
		expected     string
		errorMatcher func(error) bool
	}{
		{
			description: "basic case, 0 in last octect",
			input:       "172.31.0.0/16",
			expected:    "172.31.0.10",
		},
		{
			description:  "error, not a CIDR block",
			input:        "134.200.12.0",
			errorMatcher: IsInvalidConfig,
		},
		{
			description:  "error, last octect != 0",
			input:        "134.200.12.91/24",
			errorMatcher: IsInvalidConfig,
		},
		{
			description:  "error, not an ipv4 ip",
			input:        "not-an-actual-ip",
			errorMatcher: IsInvalidConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			actual, err := DNSIP(tc.input)

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

			if actual != tc.expected {
				t.Fatalf("DNSIP %#q doesn't match expected %#q", actual, tc.expected)
			}
		})
	}
}
