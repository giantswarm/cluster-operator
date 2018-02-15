package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func Test_ClusterID(t *testing.T) {
	testCases := []struct {
		CustomObject v1alpha1.KVMClusterConfig
		ExpectedID   string
	}{
		{
			CustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedID:   "",
		},
		{
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

	for i, tc := range testCases {
		clusterID := ClusterID(tc.CustomObject)
		if clusterID != tc.ExpectedID {
			t.Errorf("TestCase %d: ClusterID %s doesn't match. Expected: %s", (i + 1), clusterID, tc.ExpectedID)
		}
	}
}

func Test_EncryptionKeySecretName(t *testing.T) {
	testCases := []struct {
		CustomObject       v1alpha1.KVMClusterConfig
		ExpectedSecretName string
	}{
		{
			CustomObject:       v1alpha1.KVMClusterConfig{},
			ExpectedSecretName: "-encryption",
		},
		{
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

	for i, tc := range testCases {
		encryptionKeySecretName := EncryptionKeySecretName(tc.CustomObject)
		if encryptionKeySecretName != tc.ExpectedSecretName {
			t.Errorf("TestCase %d: EncryptionKeySecretName %s doesn't match. Expected: %s",
				(i + 1), encryptionKeySecretName, tc.ExpectedSecretName)
		}
	}

}

func Test_ToCustomObject(t *testing.T) {
	var emptyKVMClusterConfigPtr *v1alpha1.KVMClusterConfig

	testCases := []struct {
		InputObject          interface{}
		ExpectedCustomObject v1alpha1.KVMClusterConfig
		ExpectedError        error
	}{
		// Success case - pass empty KVMClusterConfig reference
		{
			InputObject:          &v1alpha1.KVMClusterConfig{},
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedError:        nil,
		},
		// Success case - verify that internal fields are transferred as well
		{
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
		// Failure : Nil *v1alpha1.KVMClusterConfig is not accepted.
		{
			InputObject:          emptyKVMClusterConfigPtr,
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedError:        emptyValueError,
		},
		// Failure: Non-reference value of KVMClusterConfig is not accepted.
		{
			InputObject:          v1alpha1.KVMClusterConfig{},
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedError:        wrongTypeError,
		},
		// Failure: Wrong type must not be accepted
		{
			InputObject:          &v1alpha1.AzureClusterConfig{},
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedError:        wrongTypeError,
		},
		// Failure: Nil interface{} must fail
		{
			InputObject:          nil,
			ExpectedCustomObject: v1alpha1.KVMClusterConfig{},
			ExpectedError:        wrongTypeError,
		},
	}

	for i, tc := range testCases {
		customObject, err := ToCustomObject(tc.InputObject)
		if microerror.Cause(err) != tc.ExpectedError {
			t.Errorf("TestCase %d: Received error %#v doesn't match expected %#v",
				(i + 1), err, tc.ExpectedError)
		}

		if !reflect.DeepEqual(customObject, tc.ExpectedCustomObject) {
			t.Errorf("TestCase %d: CustomObject %#v doesn't match expected %#v",
				(i + 1), customObject, tc.ExpectedCustomObject)
		}
	}
}
