package key

import (
	"fmt"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
)

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
