package clusterspec

import (
	"reflect"
	"testing"

	v1alpha1core "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	v1alpha1provider "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func Test_ClusterSpec_Factory_Construction(t *testing.T) {
	testCases := []struct {
		description  string
		baseCluster  *v1alpha1provider.Cluster
		errorMatcher func(error) bool
	}{}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			_, err := NewFactory(tt.baseCluster)

			switch {
			case err == nil && tt.errorMatcher == nil: // correct; carry on
			case err != nil && tt.errorMatcher != nil:
				if !tt.errorMatcher(err) {
					t.Errorf("received error doesn't match expected one - got: %#v", err)
				}
			case err != nil && tt.errorMatcher == nil:
				t.Errorf("got unexpected error: %#v", err)
			case err == nil && tt.errorMatcher != nil:
				t.Error("expected error but didn't receive one")
			}
		})
	}
}

func Test_ClusterSpec_Construction(t *testing.T) {
	testCases := []struct {
		description        string
		baseCluster        *v1alpha1provider.Cluster
		clusterGuestConfig v1alpha1core.ClusterGuestConfig
		expectedCluster    v1alpha1provider.Cluster
		errorMatcher       func(error) bool
	}{}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			factory, err := NewFactory(tt.baseCluster)
			if err != nil {
				t.Fatalf("ClusterSpecFactory construction failed: %#v", err)
			}

			clusterSpec, err := factory.New(tt.clusterGuestConfig)

			switch {
			case err == nil && tt.errorMatcher == nil: // correct; carry on
			case err != nil && tt.errorMatcher != nil:
				if !tt.errorMatcher(err) {
					t.Errorf("received error doesn't match expected one - got: %#v", err)
				}
			case err != nil && tt.errorMatcher == nil:
				t.Errorf("got unexpected error: %#v", err)
			case err == nil && tt.errorMatcher != nil:
				t.Error("expected error but didn't receive one")
			}

			if !reflect.DeepEqual(clusterSpec, tt.expectedCluster) {
				t.Errorf("clusterSpec doesn't match; got %#v expected %#v", clusterSpec, tt.expectedCluster)
			}
		})
	}
}
