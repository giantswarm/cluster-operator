package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func Test_ClusterID(t *testing.T) {
	testCases := []struct {
		description  string
		customObject v1alpha1.KVMClusterConfig
		expectedID   string
	}{
		{
			description:  "empty value KVMClusterConfig produces empty ID",
			customObject: v1alpha1.KVMClusterConfig{},
			expectedID:   "",
		},
		{
			description: "present ID value returned as ClusterID",
			customObject: v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID: "cluster-1",
						},
					},
				},
			},
			expectedID: "cluster-1",
		},
		{
			description: "only present ID value returned as ClusterID",
			customObject: v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID:   "cluster-123",
							Name: "First cluster",
						},
					},
				},
			},
			expectedID: "cluster-123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			clusterID := ClusterID(tc.customObject)
			if clusterID != tc.expectedID {
				t.Fatalf("ClusterID %s doesn't match. expected: %s", clusterID, tc.expectedID)
			}
		})
	}
}

func Test_EncryptionKeySecretName(t *testing.T) {
	testCases := []struct {
		description        string
		customObject       v1alpha1.KVMClusterConfig
		expectedSecretName string
	}{
		{
			description:        "empty value KVMClusterConfig returns only static part of secret name",
			customObject:       v1alpha1.KVMClusterConfig{},
			expectedSecretName: "-encryption",
		},
		{
			description: "composed secret name returned when cluster ID defined in KVMClusterConfig",
			customObject: v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID: "cluster-1",
						},
					},
				},
			},
			expectedSecretName: "cluster-1-encryption",
		},
		{
			description: "only cluster ID used to compose secret name",
			customObject: v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID:   "cluster-123",
							Name: "First cluster",
						},
					},
				},
			},
			expectedSecretName: "cluster-123-encryption",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			encryptionKeySecretName := EncryptionKeySecretName(tc.customObject)
			if encryptionKeySecretName != tc.expectedSecretName {
				t.Fatalf("EncryptionKeySecretName %s doesn't match. expected: %s",
					encryptionKeySecretName, tc.expectedSecretName)
			}
		})
	}

}

func Test_ToCustomObject(t *testing.T) {
	var emptyKVMClusterConfigPtr *v1alpha1.KVMClusterConfig

	testCases := []struct {
		description          string
		inputObject          interface{}
		expectedCustomObject v1alpha1.KVMClusterConfig
		expectedError        error
	}{
		{
			description:          "reference to empty value KVMClusterConfig returns empty KVMClusterConfig",
			inputObject:          &v1alpha1.KVMClusterConfig{},
			expectedCustomObject: v1alpha1.KVMClusterConfig{},
			expectedError:        nil,
		},
		{
			description: "verify that internal KVMClusterConfig fields are returned as well",
			inputObject: &v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID:   "cluster-1",
							Name: "My own snowflake cluster",
						},
					},
				},
			},
			expectedCustomObject: v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID:   "cluster-1",
							Name: "My own snowflake cluster",
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			description:          "nil pointer to KVMClusterConfig must return emptyValueError",
			inputObject:          emptyKVMClusterConfigPtr,
			expectedCustomObject: v1alpha1.KVMClusterConfig{},
			expectedError:        emptyValueError,
		},
		{
			description:          "non-pointer value of KVMClusterConfig must return wrontTypeError",
			inputObject:          v1alpha1.KVMClusterConfig{},
			expectedCustomObject: v1alpha1.KVMClusterConfig{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "wrong type must return wrongTypeError",
			inputObject:          &v1alpha1.AzureClusterConfig{},
			expectedCustomObject: v1alpha1.KVMClusterConfig{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "nil interface{} must return wrongTypeError",
			inputObject:          nil,
			expectedCustomObject: v1alpha1.KVMClusterConfig{},
			expectedError:        wrongTypeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			customObject, err := ToCustomObject(tc.inputObject)
			if microerror.Cause(err) != tc.expectedError {
				t.Errorf("Received error %#v doesn't match expected %#v",
					err, tc.expectedError)
			}

			if !reflect.DeepEqual(customObject, tc.expectedCustomObject) {
				t.Fatalf("customObject %#v doesn't match expected %#v",
					customObject, tc.expectedCustomObject)
			}
		})
	}
}
