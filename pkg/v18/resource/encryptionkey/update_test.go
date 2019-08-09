package encryptionkey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_newUpdateChange_Does_Not_Return_Change(t *testing.T) {
	testCases := []struct {
		description        string
		clusterGuestConfig v1alpha1.ClusterGuestConfig
		currentState       interface{}
		desiredState       interface{}
		expectedSecret     *v1.Secret
		expectedError      error
	}{
		{
			description:        "encryption key secret doesn't exist yet - no updates must be created",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			currentState:       nil,
			desiredState:       newEncryptionSecret(t, "cluster-1", map[string]string{}),
			expectedSecret:     nil,
			expectedError:      nil,
		},
		{
			description:        "encryption key secret already exists - no updates must be created",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			currentState:       newEncryptionSecret(t, "cluster-1", map[string]string{}),
			desiredState:       newEncryptionSecret(t, "cluster-1", map[string]string{}),
			expectedSecret:     nil,
			expectedError:      nil,
		},
	}

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			r, err := New(Config{
				K8sClient: fake.NewSimpleClientset(),
				Logger:    logger,
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

			secret, err := r.newUpdateChange(context.TODO(), tc.clusterGuestConfig, tc.currentState, tc.desiredState)
			if microerror.Cause(err) != tc.expectedError {
				t.Fatalf("Unexpected error returned: %#v, expected %#v", err, tc.expectedError)
			}

			assertSecret(t, secret, tc.expectedSecret)
		})
	}
}
