package encryptionkey

import (
	"context"
	"errors"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

var unknownAPIError = errors.New("Unknown error from k8s API")

func Test_GetCurrentState_Reads_Secrets_For_Relevant_ClusterID(t *testing.T) {
	testCases := []struct {
		Description    string
		CustomObject   *v1alpha1.KVMClusterConfig
		PresentSecrets []*v1.Secret
		APIReactors    []k8stesting.Reactor
		ExpectedSecret *v1.Secret
		ExpectedError  error
	}{
		{
			Description:  "three clusters exist - return secret for the one where custom object belongs",
			CustomObject: newCustomObject("cluster-2"),
			PresentSecrets: []*v1.Secret{
				newEncryptionSecret(t, "cluster-1", make(map[string]string)),
				newEncryptionSecret(t, "cluster-2", make(map[string]string)),
				newEncryptionSecret(t, "cluster-3", make(map[string]string)),
			},
			APIReactors:    []k8stesting.Reactor{},
			ExpectedSecret: newEncryptionSecret(t, "cluster-2", make(map[string]string)),
			ExpectedError:  nil,
		},
		{
			Description:    "no clusters exist - return empty list of secrets",
			CustomObject:   newCustomObject("cluster-1"),
			PresentSecrets: []*v1.Secret{},
			APIReactors:    []k8stesting.Reactor{},
			ExpectedSecret: nil,
			ExpectedError:  nil,
		},
		{
			Description:  "three clusters exist - return secrets for them despite custom object referring to new one",
			CustomObject: newCustomObject("cluster-4"),
			PresentSecrets: []*v1.Secret{
				newEncryptionSecret(t, "cluster-1", make(map[string]string)),
				newEncryptionSecret(t, "cluster-2", make(map[string]string)),
				newEncryptionSecret(t, "cluster-3", make(map[string]string)),
			},
			APIReactors:    []k8stesting.Reactor{},
			ExpectedSecret: nil,
			ExpectedError:  nil,
		},
		{
			Description:    "handle unknown error returned from Kubernetes API client",
			CustomObject:   newCustomObject("cluster-4"),
			PresentSecrets: []*v1.Secret{},
			APIReactors: []k8stesting.Reactor{
				alwaysReactWithError(unknownAPIError),
			},
			ExpectedSecret: nil,
			ExpectedError:  unknownAPIError,
		},
	}

	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tc.PresentSecrets))
			for _, s := range tc.PresentSecrets {
				objs = append(objs, s)
			}

			client := fake.NewSimpleClientset(objs...)
			client.ReactionChain = append(tc.APIReactors, client.ReactionChain...)

			r, err := New(Config{
				K8sClient: client,
				Logger:    logger,
			})

			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}

			state, err := r.GetCurrentState(context.TODO(), tc.CustomObject)
			if microerror.Cause(err) != tc.ExpectedError {
				t.Fatalf("GetCurrentState() returned error %#v - expected: %#v", err, tc.ExpectedError)
			}

			if state == nil && tc.ExpectedSecret == nil {
				// Ok
				return
			}

			secret, ok := state.(*v1.Secret)
			if !ok {
				t.Fatalf("GetCurrentState() returned wrong type %T for current state. Expected %T", state, secret)
			}

			if tc.ExpectedSecret.Labels[randomkeytpr.ClusterIDLabel] != secret.Labels[randomkeytpr.ClusterIDLabel] {
				t.Fatalf("Expected secret with cluster ID label %s, found %s",
					tc.ExpectedSecret.Labels[randomkeytpr.ClusterIDLabel],
					secret.Labels[randomkeytpr.ClusterIDLabel],
				)
			}
		})
	}
}

type errorReactor struct {
	err error
}

func (e *errorReactor) Handles(_ k8stesting.Action) bool {
	return true
}

func (e *errorReactor) React(_ k8stesting.Action) (bool, runtime.Object, error) {
	return true, nil, e.err
}

func alwaysReactWithError(err error) k8stesting.Reactor {
	return &errorReactor{
		err: err,
	}
}
