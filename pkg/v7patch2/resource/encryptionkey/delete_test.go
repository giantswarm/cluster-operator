package encryptionkey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/cluster-operator/pkg/v7patch2/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func Test_ApplyDeleteChange(t *testing.T) {
	testCases := []struct {
		description         string
		clusterGuestConfig  v1alpha1.ClusterGuestConfig
		deleteChange        interface{}
		apiReactorFactories []apiReactorFactory
		expectedError       error
	}{
		{
			description:        "delete given secret",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			deleteChange:       newEncryptionSecret(t, "cluster-1", make(map[string]string)),
			apiReactorFactories: []apiReactorFactory{func(t *testing.T) k8stesting.Reactor {
				return verifySecretDeletedReactor(t, newEncryptionSecret(t, "cluster-1", make(map[string]string)))
			}},
			expectedError: nil,
		},
		{
			description:        "handle error returned by Kubernetes API client while deleting given secret",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			deleteChange:       newEncryptionSecret(t, "cluster-1", make(map[string]string)),
			apiReactorFactories: []apiReactorFactory{func(t *testing.T) k8stesting.Reactor {
				return alwaysReturnErrorReactor(unknownAPIError)
			}},
			expectedError: unknownAPIError,
		},
		{
			description:        "handle nil passed as deleteChange",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			deleteChange:       nil,
			apiReactorFactories: []apiReactorFactory{func(t *testing.T) k8stesting.Reactor {
				return alwaysReturnErrorReactor(forbiddenAPICall)
			}},
			expectedError: nil,
		},
		{
			description:        "handle nil *v1.Secret passed as deleteChange",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			deleteChange:       emptySecretPointer,
			apiReactorFactories: []apiReactorFactory{func(t *testing.T) k8stesting.Reactor {
				return alwaysReturnErrorReactor(forbiddenAPICall)
			}},
			expectedError: nil,
		},
		{
			description:         "handle wrong type value passed as deleteChange",
			clusterGuestConfig:  newClusterGuestConfig("cluster-1"),
			deleteChange:        &v1.Pod{},
			apiReactorFactories: []apiReactorFactory{},
			expectedError:       wrongTypeError,
		},
		{
			description:        "handle deletion of non-existing secret gracefully",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			deleteChange:       newEncryptionSecret(t, "cluster-1", make(map[string]string)),
			apiReactorFactories: []apiReactorFactory{func(t *testing.T) k8stesting.Reactor {
				return alwaysReturnErrorReactor(apierrors.NewNotFound(schema.GroupResource{
					Group:    "core.giantswarm.io",
					Resource: "AWSClusterConfig",
				}, key.EncryptionKeySecretName(newClusterGuestConfig("cluster-1"))))
			}},
			expectedError: nil,
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
				ToClusterObjectMetaFunc: func(v interface{}) (apismetav1.ObjectMeta, error) {
					return apismetav1.ObjectMeta{
						Namespace: v1.NamespaceDefault,
					}, nil
				},
			})

			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}

			err = r.ApplyDeleteChange(context.TODO(), tc.clusterGuestConfig, tc.deleteChange)

			if microerror.Cause(err) != tc.expectedError {
				t.Fatalf("Unexpected error returned %#v - expected: %#v", err, tc.expectedError)
			}

			// Verification of value delting is implemented in
			// verifySecretDeletedReactor implementation.
		})
	}
}
