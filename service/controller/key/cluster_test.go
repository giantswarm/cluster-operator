package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/microerror"
	clusterv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

func Test_ToCluster(t *testing.T) {
	testCases := []struct {
		description          string
		inputObject          interface{}
		expectedCustomObject clusterv1alpha2.Cluster
		expectedError        error
	}{
		{
			description:          "reference to empty value Cluster returns empty Cluster",
			inputObject:          &clusterv1alpha2.Cluster{},
			expectedCustomObject: clusterv1alpha2.Cluster{},
			expectedError:        nil,
		},
		{
			description:          "non-pointer value of Cluster must return wrongTypeError",
			inputObject:          clusterv1alpha2.Cluster{},
			expectedCustomObject: clusterv1alpha2.Cluster{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "wrong type must return wrongTypeError",
			inputObject:          &clusterv1alpha2.Machine{},
			expectedCustomObject: clusterv1alpha2.Cluster{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "nil interface{} must return wrongTypeError",
			inputObject:          nil,
			expectedCustomObject: clusterv1alpha2.Cluster{},
			expectedError:        wrongTypeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			object, err := ToCluster(tc.inputObject)
			if microerror.Cause(err) != tc.expectedError {
				t.Errorf("Received error %#v doesn't match expected %#v",
					err, tc.expectedError)
			}

			if !reflect.DeepEqual(object, tc.expectedCustomObject) {
				t.Fatalf("object %#v doesn't match expected %#v",
					object, tc.expectedCustomObject)
			}
		})
	}
}
