package key

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
)

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
