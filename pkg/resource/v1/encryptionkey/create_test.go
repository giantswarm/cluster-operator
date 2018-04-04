package encryptionkey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func Test_ApplyCreateChange(t *testing.T) {
	testCases := []struct {
		description         string
		clusterGuestConfig  v1alpha1.ClusterGuestConfig
		createChange        interface{}
		apiReactorFactories []apiReactorFactory
		expectedError       error
	}{
		{
			description:        "create given secret",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			createChange:       newEncryptionSecret(t, "cluster-1", make(map[string]string)),
			apiReactorFactories: []apiReactorFactory{func(t *testing.T) k8stesting.Reactor {
				return verifySecretCreatedReactor(t, newEncryptionSecret(t, "cluster-1", make(map[string]string)))
			}},
			expectedError: nil,
		},
		{
			description:        "handle error returned by Kubernetes API client while creating given secret",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			createChange:       newEncryptionSecret(t, "cluster-1", make(map[string]string)),
			apiReactorFactories: []apiReactorFactory{func(t *testing.T) k8stesting.Reactor {
				return alwaysReturnErrorReactor(unknownAPIError)
			}},
			expectedError: unknownAPIError,
		},
		{
			description:        "handle nil passed as createChange",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			createChange:       nil,
			apiReactorFactories: []apiReactorFactory{func(t *testing.T) k8stesting.Reactor {
				return alwaysReturnErrorReactor(forbiddenAPICall)
			}},
			expectedError: nil,
		},
		{
			description:        "handle nil *v1.Secret passed as createChange",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			createChange:       emptySecretPointer,
			apiReactorFactories: []apiReactorFactory{func(t *testing.T) k8stesting.Reactor {
				return alwaysReturnErrorReactor(forbiddenAPICall)
			}},
			expectedError: nil,
		},
		{
			description:         "handle wrong type value passed as createChange",
			clusterGuestConfig:  newClusterGuestConfig("cluster-1"),
			createChange:        &v1.Pod{},
			apiReactorFactories: []apiReactorFactory{},
			expectedError:       wrongTypeError,
		},
	}

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			client := fake.NewSimpleClientset()

			// Reactor order matters - hence construction to intermediate slice.
			apiReactors := make([]k8stesting.Reactor, 0, len(tc.apiReactorFactories))
			for _, factory := range tc.apiReactorFactories {
				apiReactors = append(apiReactors, factory(t))
			}

			// Prepend test reactors before existing ones because reactors are executed in order
			client.ReactionChain = append(apiReactors, client.ReactionChain...)

			r, err := New(Config{
				K8sClient:   client,
				Logger:      logger,
				ProjectName: "cluster-operator",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
			})

			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}

			err = r.ApplyCreateChange(context.TODO(), tc.clusterGuestConfig, tc.createChange)

			if microerror.Cause(err) != tc.expectedError {
				t.Fatalf("Unexpected error returned %#v - expected: %#v", err, tc.expectedError)
			}

			// Created value verification is implemented in k8stesting.Reactor
			// implementation.
		})
	}
}

func Test_newCreateChange(t *testing.T) {
	testCases := []struct {
		description        string
		clusterGuestConfig v1alpha1.ClusterGuestConfig
		currentState       interface{}
		desiredState       interface{}
		expectedSecret     *v1.Secret
		expectedError      error
	}{
		{
			description:        "encryption key secret doesn't exist yet - secret should create it",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			currentState:       nil,
			desiredState:       newEncryptionSecret(t, "cluster-1", map[string]string{}),
			expectedSecret:     newEncryptionSecret(t, "cluster-1", map[string]string{}),
			expectedError:      nil,
		},
		{
			description:        "encryption key secret already exists - secret must not be created",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			currentState:       newEncryptionSecret(t, "cluster-1", map[string]string{}),
			desiredState:       newEncryptionSecret(t, "cluster-1", map[string]string{}),
			expectedSecret:     nil,
			expectedError:      nil,
		},
		{
			description:        "verify currentState type verification error handling",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			currentState:       &v1.Pod{},
			desiredState:       newEncryptionSecret(t, "cluster-1", map[string]string{}),
			expectedSecret:     nil,
			expectedError:      wrongTypeError,
		},
		{
			description:        "verify desiredState type verification error handling",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			currentState:       nil,
			desiredState:       &v1.Pod{},
			expectedSecret:     nil,
			expectedError:      wrongTypeError,
		},
	}

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			r, err := New(Config{
				K8sClient:   fake.NewSimpleClientset(),
				Logger:      logger,
				ProjectName: "cluster-operator",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
			})

			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}

			secret, err := r.newCreateChange(context.TODO(), tc.clusterGuestConfig, tc.currentState, tc.desiredState)
			if microerror.Cause(err) != tc.expectedError {
				t.Fatalf("Unexpected error returned: %#v, expected %#v", err, tc.expectedError)
			}

			assertSecret(t, secret, tc.expectedSecret)
		})
	}
}
