package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func Test_ToCustomObject(t *testing.T) {
	var emptyClusterNetworkConfigPtr *v1alpha1.ClusterNetworkConfig

	testCases := []struct {
		description          string
		inputObject          interface{}
		expectedCustomObject v1alpha1.ClusterNetworkConfig
		expectedError        error
	}{
		{
			description:          "reference to empty value ClusterNetworkConfig returns empty ClusterNetworkConfig",
			inputObject:          &v1alpha1.ClusterNetworkConfig{},
			expectedCustomObject: v1alpha1.ClusterNetworkConfig{},
			expectedError:        nil,
		},
		{
			description: "verify that internal ClusterNetworkConfig fields are returned as well",
			inputObject: &v1alpha1.ClusterNetworkConfig{
				Spec: v1alpha1.ClusterNetworkConfigSpec{
					MaskBits: 31,
				},
			},
			expectedCustomObject: v1alpha1.ClusterNetworkConfig{
				Spec: v1alpha1.ClusterNetworkConfigSpec{
					MaskBits: 31,
				},
			},
			expectedError: nil,
		},
		{
			description:          "nil pointer to ClusterNetworkConfig must return emptyValueError",
			inputObject:          emptyClusterNetworkConfigPtr,
			expectedCustomObject: v1alpha1.ClusterNetworkConfig{},
			expectedError:        emptyValueError,
		},
		{
			description:          "non-pointer value of ClusterNetworkConfig must return wrontTypeError",
			inputObject:          v1alpha1.ClusterNetworkConfig{},
			expectedCustomObject: v1alpha1.ClusterNetworkConfig{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "wrong type must return wrongTypeError",
			inputObject:          &v1alpha1.AzureClusterConfig{},
			expectedCustomObject: v1alpha1.ClusterNetworkConfig{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "nil interface{} must return wrongTypeError",
			inputObject:          nil,
			expectedCustomObject: v1alpha1.ClusterNetworkConfig{},
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
