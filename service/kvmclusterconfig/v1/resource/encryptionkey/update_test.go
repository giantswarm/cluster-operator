package encryptionkey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_NewUpdatePatch_Computes_Patch_Correctly(t *testing.T) {
	testCases := []struct {
		CustomObject  *v1alpha1.KVMClusterConfig
		CurrentState  interface{}
		DesiredState  interface{}
		ExpectedPatch *framework.Patch
		ExpectedError error
	}{
		// Encryption key secret doesn't exist yet.
		// Patch should create it.
		{
			CustomObject: createCustomObject("cluster-1"),
			CurrentState: nil,
			DesiredState: createEncryptionSecret(t, "cluster-1", map[string]string{}),
			ExpectedPatch: func() *framework.Patch {
				p := framework.NewPatch()
				p.SetCreateChange(createEncryptionSecret(t, "cluster-1", map[string]string{}))
				return p
			}(),
			ExpectedError: nil,
		},
		// Encryption key secret already exists.
		// Patch must not create it.
		{
			CustomObject:  createCustomObject("cluster-1"),
			CurrentState:  createEncryptionSecret(t, "cluster-1", map[string]string{}),
			DesiredState:  createEncryptionSecret(t, "cluster-1", map[string]string{}),
			ExpectedPatch: framework.NewPatch(),
			ExpectedError: nil,
		},
		// Verify currentState type verification error handling
		{
			CustomObject:  createCustomObject("cluster-1"),
			CurrentState:  &v1.Pod{},
			DesiredState:  createEncryptionSecret(t, "cluster-1", map[string]string{}),
			ExpectedPatch: nil,
			ExpectedError: wrongTypeError,
		},
		// Verify desiredState type verification error handling
		{
			CustomObject:  createCustomObject("cluster-1"),
			CurrentState:  nil,
			DesiredState:  &v1.Pod{},
			ExpectedPatch: nil,
			ExpectedError: wrongTypeError,
		},
	}

	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for i, tc := range testCases {
		r, err := New(Config{
			K8sClient: fake.NewSimpleClientset(),
			Logger:    logger,
		})

		patch, err := r.NewUpdatePatch(context.TODO(), tc.CustomObject, tc.CurrentState, tc.DesiredState)
		if microerror.Cause(err) != tc.ExpectedError {
			t.Errorf("TestCase %d: Unexpected error returned: %#v, expected %#v", (i + 1), err, tc.ExpectedError)
			continue
		}

		assertPatch(t, (i + 1), patch, tc.ExpectedPatch)
	}
}
