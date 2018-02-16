package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func Test_ClusterID(t *testing.T) {
	testCases := []struct {
		Description  string
		CustomObject v1alpha1.KVMClusterConfig
		ExpectedID   string
	}{
		{
			Description:  "empty value KVMClusterConfig produces empty ID",
			CustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedID:   "",
		},
		{
			Description: "present ID value returned as ClusterID",
			CustomObject: v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID: "cluster-1",
						},
					},
				},
			},
			ExpectedID: "cluster-1",
		},
		{
			Description: "only present ID value returned as ClusterID",
			CustomObject: v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID:   "cluster-123",
							Name: "First cluster",
						},
					},
				},
			},
			ExpectedID: "cluster-123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			clusterID := ClusterID(tc.CustomObject)
			if clusterID != tc.ExpectedID {
				t.Fatalf("ClusterID %s doesn't match. Expected: %s", clusterID, tc.ExpectedID)
			}
		})
	}
}

func Test_EncryptionKeySecretName(t *testing.T) {
	testCases := []struct {
		Description        string
		CustomObject       v1alpha1.KVMClusterConfig
		ExpectedSecretName string
	}{
		{
			Description:        "empty value KVMClusterConfig returns only static part of secret name",
			CustomObject:       v1alpha1.KVMClusterConfig{},
			ExpectedSecretName: "-encryption",
		},
		{
			Description: "composed secret name returned when cluster ID defined in KVMClusterConfig",
			CustomObject: v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID: "cluster-1",
						},
					},
				},
			},
			ExpectedSecretName: "cluster-1-encryption",
		},
		{
			Description: "only cluster ID used to compose secret name",
			CustomObject: v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID:   "cluster-123",
							Name: "First cluster",
						},
					},
				},
			},
			ExpectedSecretName: "cluster-123-encryption",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			encryptionKeySecretName := EncryptionKeySecretName(tc.CustomObject)
			if encryptionKeySecretName != tc.ExpectedSecretName {
				t.Fatalf("EncryptionKeySecretName %s doesn't match. Expected: %s",
					encryptionKeySecretName, tc.ExpectedSecretName)
			}
		})
	}

}

func Test_ToCustomObject(t *testing.T) {
	var emptyKVMClusterConfigPtr *v1alpha1.KVMClusterConfig

	testCases := []struct {
		Description          string
		InputObject          interface{}
		ExpectedCustomObject v1alpha1.KVMClusterConfig
		ExpectedError        error
	}{
		{
			Description:          "reference to empty value KVMClusterConfig returns empty KVMClusterConfig",
			InputObject:          &v1alpha1.KVMClusterConfig{},
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedError:        nil,
		},
		{
			Description: "verify that internal KVMClusterConfig fields are returned as well",
			InputObject: &v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID:   "cluster-1",
							Name: "My own snowflake cluster",
						},
					},
				},
			},
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{
				Spec: v1alpha1.KVMClusterConfigSpec{
					Guest: v1alpha1.KVMClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID:   "cluster-1",
							Name: "My own snowflake cluster",
						},
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Description:          "nil pointer to KVMClusterConfig must return emptyValueError",
			InputObject:          emptyKVMClusterConfigPtr,
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedError:        emptyValueError,
		},
		{
			Description:          "non-pointer value of KVMClusterConfig must return wrontTypeError",
			InputObject:          v1alpha1.KVMClusterConfig{},
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedError:        wrongTypeError,
		},
		{
			Description:          "wrong type must return wrongTypeError",
			InputObject:          &v1alpha1.AzureClusterConfig{},
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedError:        wrongTypeError,
		},
		{
			Description:          "nil interface{} must return wrongTypeError",
			InputObject:          nil,
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedError:        wrongTypeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			customObject, err := ToCustomObject(tc.InputObject)
			if microerror.Cause(err) != tc.ExpectedError {
				t.Errorf("Received error %#v doesn't match expected %#v",
					err, tc.ExpectedError)
			}

			if !reflect.DeepEqual(customObject, tc.ExpectedCustomObject) {
				t.Fatalf("CustomObject %#v doesn't match expected %#v",
					customObject, tc.ExpectedCustomObject)
			}
		})
	}
}
