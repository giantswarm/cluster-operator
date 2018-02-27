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
			name := CertConfigName(tc.clusterConfig, tc.cert)
			if name != tc.expectedCertConfigName {
				t.Fatalf("CertConfigName '%s' doesn't match expected '%s'", name, tc.expectedCertConfigName)
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
