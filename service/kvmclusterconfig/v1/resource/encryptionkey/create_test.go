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

func Test_newCreateChange(t *testing.T) {
	testCases := []struct {
		Description    string
		CustomObject   *v1alpha1.KVMClusterConfig
		CurrentState   interface{}
		DesiredState   interface{}
		ExpectedSecret *v1.Secret
		ExpectedError  error
	}{
		{
			Description:    "encryption key secret doesn't exist yet - secret should create it",
			CustomObject:   newCustomObject("cluster-1"),
			CurrentState:   nil,
			DesiredState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			ExpectedSecret: newEncryptionSecret(t, "cluster-1", map[string]string{}),
			ExpectedError:  nil,
		},
		{
			Description:    "encryption key secret already exists - secret must not be created",
			CustomObject:   newCustomObject("cluster-1"),
			CurrentState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			DesiredState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			ExpectedSecret: nil,
			ExpectedError:  nil,
		},
		{
			Description:    "verify currentState type verification error handling",
			CustomObject:   newCustomObject("cluster-1"),
			CurrentState:   &v1.Pod{},
			DesiredState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			ExpectedSecret: nil,
			ExpectedError:  wrongTypeError,
		},
		{
			Description:    "verify desiredState type verification error handling",
			CustomObject:   newCustomObject("cluster-1"),
			CurrentState:   nil,
			DesiredState:   &v1.Pod{},
			ExpectedSecret: nil,
			ExpectedError:  wrongTypeError,
		},
	}

	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			r, err := New(Config{
				K8sClient: fake.NewSimpleClientset(),
				Logger:    logger,
			})

			secret, err := r.newCreateChange(context.TODO(), tc.CustomObject, tc.CurrentState, tc.DesiredState)
			if microerror.Cause(err) != tc.ExpectedError {
				t.Fatalf("Unexpected error returned: %#v, expected %#v", err, tc.ExpectedError)
			}

			assertSecret(t, secret, tc.ExpectedSecret)
		})
	}
}
