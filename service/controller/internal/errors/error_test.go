package errors

import (
	"errors"
	"strconv"
	"testing"
)

func Test_IsChartConfigNotAvailable(t *testing.T) {
	testCases := []struct {
		name          string
		errorMessage  string
		expectedMatch bool
	}{
		{
			name:          "case 0: chartconfig not ready get EOF error",
			errorMessage:  "Get https://api.y2e65.k8s.geckon.gridscale.kvm.gigantic.io/apis/core.giantswarm.io/v1alpha1/namespaces/giantswarm/chartconfigs?labelSelector=giantswarm.io%2Fmanaged-by%3Dcluster-operator: EOF",
			expectedMatch: true,
		},
		{
			name:          "case 1: chartconfig request canceled",
			errorMessage:  "Get https://api.q6irk.k8s.geckon.gridscale.kvm.gigantic.io/apis/core.giantswarm.io/v1alpha1/namespaces/giantswarm/chartconfigs?labelSelector=giantswarm.io%2Fmanaged-by%3Dcluster-operator: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)",
			expectedMatch: true,
		},
		{
			name:          "case 2: nodes EOF error does not match",
			errorMessage:  "Get https://api.5xchu.aws.gigantic.io/api/v1/nodes: EOF",
			expectedMatch: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			err := errors.New(tc.errorMessage)
			result := IsChartConfigNotAvailable(err)

			if result != tc.expectedMatch {
				t.Fatalf("expected %t, got %t", tc.expectedMatch, result)
			}
		})
	}
}
