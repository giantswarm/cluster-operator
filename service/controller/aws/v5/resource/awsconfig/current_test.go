package awsconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

func Test_GetCurrentState(t *testing.T) {
	testCases := []struct {
		name              string
		awsClusterConfig  *v1alpha1.AWSClusterConfig
		presentAWSConfigs []*providerv1alpha1.AWSConfig
		apiReactors       []k8stesting.Reactor
		expectedAWSConfig *providerv1alpha1.AWSConfig
		errorMatcher      func(error) bool
	}{
		{
			name:             "case 0: three clusters exist - return AWSConfig for the one where custom object belongs",
			awsClusterConfig: newAWSClusterConfig("abcd3e", "test cluster"),
			presentAWSConfigs: []*providerv1alpha1.AWSConfig{
				newAWSConfig("abcd3e"),
				newAWSConfig("qwer9t"),
				newAWSConfig("zxcv0b"),
			},
			apiReactors:       []k8stesting.Reactor{},
			expectedAWSConfig: newAWSConfig("abcd3e"),
			errorMatcher:      nil,
		},
		{
			name:              "case 1: no clusters exist - return empty list of secrets",
			awsClusterConfig:  newAWSClusterConfig("abcd3e", "test cluster"),
			presentAWSConfigs: []*providerv1alpha1.AWSConfig{},
			apiReactors:       []k8stesting.Reactor{},
			expectedAWSConfig: nil,
			errorMatcher:      nil,
		},
		{
			name:             "case 2: three clusters exist - none of them is for given AWSClusterConfig",
			awsClusterConfig: newAWSClusterConfig("abcd3e", "test cluster"),
			presentAWSConfigs: []*providerv1alpha1.AWSConfig{
				newAWSConfig("asdf8g"),
				newAWSConfig("qwer9t"),
				newAWSConfig("zxcv0b"),
			},
			apiReactors:       []k8stesting.Reactor{},
			expectedAWSConfig: nil,
			errorMatcher:      nil,
		},
		{
			name:              "case 3: handle unknown error returned from Kubernetes API client",
			awsClusterConfig:  newAWSClusterConfig("abcd3e", "test cluster"),
			presentAWSConfigs: []*providerv1alpha1.AWSConfig{},
			apiReactors: []k8stesting.Reactor{
				alwaysReturnErrorReactor(unknownAPIError),
			},
			expectedAWSConfig: nil,
			errorMatcher:      func(err error) bool { return microerror.Cause(err) == unknownAPIError },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tc.presentAWSConfigs))
			for _, s := range tc.presentAWSConfigs {
				objs = append(objs, s)
			}

			client := fake.NewSimpleClientset(objs...)
			client.ReactionChain = append(tc.apiReactors, client.ReactionChain...)

			r, err := New(Config{
				G8sClient: client,
				Logger:    microloggertest.New(),
			})

			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}

			state, err := r.GetCurrentState(context.TODO(), tc.awsClusterConfig)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if state == nil && tc.expectedAWSConfig == nil {
				// Ok
				return
			}

			awsConfig, ok := state.(*providerv1alpha1.AWSConfig)
			if !ok {
				t.Fatalf("GetCurrentState() returned wrong type %T for current state. expected %T", state, awsConfig)
			}

			if !reflect.DeepEqual(awsConfig, tc.expectedAWSConfig) {
				t.Fatalf("GetCurrentState() == %#v, expected %#v", awsConfig, tc.expectedAWSConfig)
			}
		})
	}
}
