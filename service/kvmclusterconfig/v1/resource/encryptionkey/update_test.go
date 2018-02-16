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

func Test_newUpdateChange_Does_Not_Return_Change(t *testing.T) {
	testCases := []struct {
		Description    string
		CustomObject   *v1alpha1.KVMClusterConfig
		CurrentState   interface{}
		DesiredState   interface{}
		ExpectedSecret *v1.Secret
		ExpectedError  error
	}{
		{
			Description:    "encryption key secret doesn't exist yet - no updates must be created",
			CustomObject:   newCustomObject("cluster-1"),
			CurrentState:   nil,
			DesiredState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			ExpectedSecret: nil,
			ExpectedError:  nil,
		},
		{
			Description:    "encryption key secret already exists - no updates must be created",
			CustomObject:   newCustomObject("cluster-1"),
			CurrentState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			DesiredState:   newEncryptionSecret(t, "cluster-1", map[string]string{}),
			ExpectedSecret: nil,
			ExpectedError:  nil,
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

			secret, err := r.newUpdateChange(context.TODO(), tc.CustomObject, tc.CurrentState, tc.DesiredState)
			if microerror.Cause(err) != tc.ExpectedError {
				t.Errorf("Unexpected error returned: %#v, expected %#v", err, tc.ExpectedError)
				return
			}

			assertSecret(t, secret, tc.ExpectedSecret)
		})
	}
}
