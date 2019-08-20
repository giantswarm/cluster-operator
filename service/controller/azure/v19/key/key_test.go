package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func Test_ToCustomObject(t *testing.T) {
	var emptyAzureClusterConfigPtr *v1alpha1.AzureClusterConfig

	testCases := []struct {
		description          string
		inputObject          interface{}
		expectedCustomObject v1alpha1.AzureClusterConfig
		expectedError        error
	}{
		{
			description:          "reference to empty value AzureClusterConfig returns empty AzureClusterConfig",
			inputObject:          &v1alpha1.AzureClusterConfig{},
			expectedCustomObject: v1alpha1.AzureClusterConfig{},
			expectedError:        nil,
		},
		{
			description: "verify that internal AzureClusterConfig fields are returned as well",
			inputObject: &v1alpha1.AzureClusterConfig{
				Spec: v1alpha1.AzureClusterConfigSpec{
					Guest: v1alpha1.AzureClusterConfigSpecGuest{
						ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
							ID:   "cluster-1",
							Name: "My own snowflake cluster",
						},
					},
				},
			},
			expectedCustomObject: v1alpha1.AzureClusterConfig{
				Spec: v1alpha1.AzureClusterConfigSpec{
					Guest: v1alpha1.AzureClusterConfigSpecGuest{
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
			description:          "nil pointer to AzureClusterConfig must return emptyValueError",
			inputObject:          emptyAzureClusterConfigPtr,
			expectedCustomObject: v1alpha1.AzureClusterConfig{},
			expectedError:        emptyValueError,
		},
		{
			description:          "non-pointer value of AzureClusterConfig must return wrongTypeError",
			inputObject:          v1alpha1.AzureClusterConfig{},
			expectedCustomObject: v1alpha1.AzureClusterConfig{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "wrong type must return wrongTypeError",
			inputObject:          &v1alpha1.AWSClusterConfig{},
			expectedCustomObject: v1alpha1.AzureClusterConfig{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "nil interface{} must return wrongTypeError",
			inputObject:          nil,
			expectedCustomObject: v1alpha1.AzureClusterConfig{},
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

func Test_WorkerCount(t *testing.T) {
	testCases := []struct {
		description         string
		clusterConfig       v1alpha1.AzureClusterConfig
		expectedWorkerCount int
	}{
		{
			description:         "case 0: empty value",
			clusterConfig:       v1alpha1.AzureClusterConfig{},
			expectedWorkerCount: 0,
		},
		{
			description: "case 1: basic match",
			clusterConfig: v1alpha1.AzureClusterConfig{
				Spec: v1alpha1.AzureClusterConfigSpec{
					Guest: v1alpha1.AzureClusterConfigSpecGuest{
						Workers: []v1alpha1.AzureClusterConfigSpecGuestWorker{
							{},
						},
					},
				},
			},
			expectedWorkerCount: 1,
		},
		{
			description: "case 2: different worker count",
			clusterConfig: v1alpha1.AzureClusterConfig{
				Spec: v1alpha1.AzureClusterConfigSpec{
					Guest: v1alpha1.AzureClusterConfigSpecGuest{
						Workers: []v1alpha1.AzureClusterConfigSpecGuestWorker{
							{},
							{},
							{},
						},
					},
				},
			},
			expectedWorkerCount: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			workerCount := WorkerCount(tc.clusterConfig)
			if workerCount != tc.expectedWorkerCount {
				t.Fatalf("WorkerCount %d doesn't match expected %d", workerCount, tc.expectedWorkerCount)
			}
		})
	}
}
