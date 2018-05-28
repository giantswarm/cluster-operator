package key

import (
	"fmt"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_APIDomain(t *testing.T) {
	testCases := []struct {
		description       string
		clusterConfig     v1alpha1.ClusterGuestConfig
		expectedAPIDomain string
		errorMatcher      func(error) bool
	}{
		{
			description: "case 0: basic match",
			clusterConfig: v1alpha1.ClusterGuestConfig{
				DNSZone: "rue99.k8s.gauss.eu-central-1.aws.gigantic.io",
			},
			expectedAPIDomain: "api.rue99.k8s.gauss.eu-central-1.aws.gigantic.io",
		},
		{
			description: "case 1: different DNSZone",
			clusterConfig: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.k8s.gollum.westeurope.azure.gigantic.io",
			},
			expectedAPIDomain: "api.5xchu.k8s.gollum.westeurope.azure.gigantic.io",
		},
		{
			description: "case 2: invalid DNSZone",
			clusterConfig: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu",
			},
			errorMatcher: IsInvalidConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			domain, err := APIDomain(tc.clusterConfig)

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

			if domain != tc.expectedAPIDomain {
				t.Fatalf("APIDomain '%s' doesn't match expected '%s'", domain, tc.expectedAPIDomain)
			}
		})
	}
}

func Test_CertConfigName(t *testing.T) {
	testCases := []struct {
		description            string
		clusterConfig          v1alpha1.ClusterGuestConfig
		cert                   certs.Cert
		expectedCertConfigName string
	}{
		{
			description:   "empty ClusterGuestConfig value with APICert",
			clusterConfig: v1alpha1.ClusterGuestConfig{},
			cert:          certs.APICert,
			expectedCertConfigName: fmt.Sprintf("-%s", certs.APICert),
		},
		{
			description: "ClusterGuestConfig with ID and WorkerCert",
			clusterConfig: v1alpha1.ClusterGuestConfig{
				ID:   "cluster-1",
				Name: "Test cluster nr. 1",
			},
			cert: certs.WorkerCert,
			expectedCertConfigName: fmt.Sprintf("cluster-1-%s", certs.WorkerCert),
		},
		{
			description: "ClusterGuestConfig with ID and empty value for cert",
			clusterConfig: v1alpha1.ClusterGuestConfig{
				ID:   "cluster-1",
				Name: "Test cluster nr. 1",
			},
			cert: "",
			expectedCertConfigName: "cluster-1-",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			name := CertConfigName(ClusterID(tc.clusterConfig), tc.cert)
			if name != tc.expectedCertConfigName {
				t.Fatalf("CertConfigName '%s' doesn't match expected '%s'", name, tc.expectedCertConfigName)
			}
		})
	}
}

func Test_CertConfigVersionBundleVersion(t *testing.T) {
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
				Spec: v1alpha1.CertConfigSpec{
					VersionBundle: v1alpha1.CertConfigSpecVersionBundle{
						Version: "1.0.1",
					},
				},
			},
			expectedVersion: "1.0.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			version := CertConfigVersionBundleVersion(tc.certConfig)
			if version != tc.expectedVersion {
				t.Fatalf("version '%s' doesn't match expected '%s'", version, tc.expectedVersion)
			}
		})
	}
}

func Test_ClusterID(t *testing.T) {
	testCases := []struct {
		description   string
		clusterConfig v1alpha1.ClusterGuestConfig
		expectedID    string
	}{
		{
			description:   "empty value",
			clusterConfig: v1alpha1.ClusterGuestConfig{},
			expectedID:    "",
		},
		{
			description: "ClusterGuestConfig with ID",
			clusterConfig: v1alpha1.ClusterGuestConfig{
				ID:   "cluster-1",
				Name: "Test cluster nr. 1",
			},
			expectedID: "cluster-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			id := ClusterID(tc.clusterConfig)
			if id != tc.expectedID {
				t.Fatalf("ClusterID '%s' doesn't match expected '%s'", id, tc.expectedID)
			}
		})
	}
}

func Test_EncryptionKeySecretName(t *testing.T) {
	testCases := []struct {
		description        string
		clusterGuestConfig v1alpha1.ClusterGuestConfig
		expectedSecretName string
	}{
		{
			description:        "empty value KVMClusterConfig returns only static part of secret name",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{},
			expectedSecretName: "-encryption",
		},
		{
			description: "composed secret name returned when cluster ID defined in KVMClusterConfig",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			expectedSecretName: "cluster-1-encryption",
		},
		{
			description: "only cluster ID used to compose secret name",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID:   "cluster-123",
				Name: "First cluster",
			},
			expectedSecretName: "cluster-123-encryption",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			encryptionKeySecretName := EncryptionKeySecretName(tc.clusterGuestConfig)
			if encryptionKeySecretName != tc.expectedSecretName {
				t.Fatalf("EncryptionKeySecretName %s doesn't match. expected: %s",
					encryptionKeySecretName, tc.expectedSecretName)
			}
		})
	}

}

func Test_IsDeleted(t *testing.T) {
	testCases := []struct {
		description    string
		objectMeta     apismetav1.ObjectMeta
		expectedResult bool
	}{
		{
			description:    "case 0: false when struct is empty",
			objectMeta:     apismetav1.ObjectMeta{},
			expectedResult: false,
		},
		{
			description: "case 1: false when field is nil",
			objectMeta: apismetav1.ObjectMeta{
				DeletionTimestamp: nil,
			},
			expectedResult: false,
		},
		{
			description: "case 2: true when field is set",
			objectMeta: apismetav1.ObjectMeta{
				DeletionTimestamp: &apismetav1.Time{Time: time.Now()},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := IsDeleted(tc.objectMeta)
			if result != tc.expectedResult {
				t.Fatalf("expected IsDeleted %t, got %t", tc.expectedResult, result)
			}
		})
	}
}

func Test_MasterServiceDomain(t *testing.T) {
	testCases := []struct {
		description    string
		clusterConfig  v1alpha1.ClusterGuestConfig
		expectedDomain string
	}{
		{
			description: "basic match",
			clusterConfig: v1alpha1.ClusterGuestConfig{
				ID: "5xchu",
			},
			expectedDomain: "master.5xchu",
		},
		{
			description: "different cluster id",
			clusterConfig: v1alpha1.ClusterGuestConfig{
				ID: "rue99",
			},
			expectedDomain: "master.rue99",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			domain := MasterServiceDomain(tc.clusterConfig)
			if domain != tc.expectedDomain {
				t.Fatalf("MasterServiceDomain '%s' doesn't match expected '%s'", domain, tc.expectedDomain)
			}
		})
	}
}

func Test_ServerDomain(t *testing.T) {
	testCases := []struct {
		description    string
		cert           certs.Cert
		clusterConfig  v1alpha1.ClusterGuestConfig
		expectedDomain string
		errorMatcher   func(error) bool
	}{
		{
			description: "case 0: basic match",
			cert:        certs.APICert,
			clusterConfig: v1alpha1.ClusterGuestConfig{
				DNSZone: "rue99.k8s.gauss.eu-central-1.aws.gigantic.io",
			},
			expectedDomain: "api.rue99.k8s.gauss.eu-central-1.aws.gigantic.io",
		},
		{
			description: "case 1: different DNSZone",
			cert:        certs.APICert,
			clusterConfig: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.k8s.gollum.westeurope.azure.gigantic.io",
			},
			expectedDomain: "api.5xchu.k8s.gollum.westeurope.azure.gigantic.io",
		},
		{
			description: "case 2: different cert",
			cert:        certs.EtcdCert,
			clusterConfig: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu.k8s.gollum.westeurope.azure.gigantic.io",
			},
			expectedDomain: "etcd.5xchu.k8s.gollum.westeurope.azure.gigantic.io",
		},
		{
			description: "case 3: invalid DNSZone",
			cert:        certs.APICert,
			clusterConfig: v1alpha1.ClusterGuestConfig{
				DNSZone: "5xchu",
			},
			errorMatcher: IsInvalidConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			domain, err := serverDomain(tc.clusterConfig, tc.cert)

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

			if domain != tc.expectedDomain {
				t.Fatalf("ServerDomain '%s' doesn't match expected '%s'", domain, tc.expectedDomain)
			}
		})
	}
}
