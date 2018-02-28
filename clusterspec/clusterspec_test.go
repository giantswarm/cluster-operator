package clusterspec

import (
	"reflect"
	"testing"

	v1alpha1core "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	v1alpha1provider "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/cluster-operator/flag"
)

func Test_ClusterSpec_Factory_Construction(t *testing.T) {
	testCases := []struct {
		description  string
		flag         *flag.Flag
		errorMatcher func(error) bool
	}{}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			_, err := NewFactory(tt.flag)

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
		clusterGuestConfig v1alpha1core.ClusterGuestConfig
		expectedCluster    v1alpha1provider.Cluster
		errorMatcher       func(error) bool
	}{}

	flag := &flag.Flag{}
	factory, err := NewFactory(flag)
	if err != nil {
		t.Fatalf("ClusterSpecFactory construction failed: %#v", err)
	}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
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
