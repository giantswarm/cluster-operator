package encryptionkey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_NewDeletePatch_Computes_Patch_Correctly(t *testing.T) {
	testCases := []struct {
		description    string
		customObject   *v1alpha1.KVMClusterConfig
		currentState   interface{}
		desiredState   interface{}
		expectedSecret *v1.Secret
		expectedError  error
	}{
		{
			description:    "delete existing secret when encryption key custom object is deleted",
			customObject:   newCustomObject("cluster-1"),
			currentState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			desiredState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			expectedSecret: newEncryptionSecret(t, "cluster-1", map[string]string{}),
			expectedError:  nil,
		},
		{
			description:    "nothing to delete when encryption key custom object get's deleted",
			customObject:   newCustomObject("cluster-1"),
			currentState:   nil,
			desiredState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			expectedSecret: nil,
			expectedError:  nil,
		},
		{
			description:    "verify currentState type verification error handling",
			customObject:   newCustomObject("cluster-1"),
			currentState:   &v1.Pod{},
			desiredState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			expectedSecret: nil,
			expectedError:  wrongTypeError,
		},
	}

	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			r, err := New(Config{
				K8sClient: fake.NewSimpleClientset(),
				Logger:    logger,
			})

			secret, err := r.newDeleteChangeForDeletePatch(context.TODO(), tc.customObject, tc.currentState, tc.desiredState)
			if microerror.Cause(err) != tc.expectedError {
				t.Fatalf("Unexpected error returned: %#v, expected %#v", err, tc.expectedError)
			}

			assertSecret(t, secret, tc.expectedSecret)
		})
	}
}
