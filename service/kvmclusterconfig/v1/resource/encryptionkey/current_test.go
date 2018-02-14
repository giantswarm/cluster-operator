package encryptionkey

import (
	"context"
	"errors"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/cluster-operator/service/kvmclusterconfig/v1/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

var unknownAPIError = errors.New("Unknown error from k8s API")

func Test_GetCurrentState_Reads_Secrets_For_Relevant_ClusterID(t *testing.T) {
	testCases := []struct {
		CustomObject   *v1alpha1.KVMClusterConfig
		PresentSecrets []*v1.Secret
		APIReactors    []k8stesting.Reactor
		ExpectedSecret *v1.Secret
		ExpectedError  error
	}{
		// Cluster exists: Success case when there are three cluster secrets
		// present and one of them is correct
		{
			CustomObject: createCustomObject("cluster-2"),
			PresentSecrets: []*v1.Secret{
				createEncryptionSecret(t, "cluster-1", make(map[string]string)),
				createEncryptionSecret(t, "cluster-2", make(map[string]string)),
				createEncryptionSecret(t, "cluster-3", make(map[string]string)),
			},
			APIReactors:    []k8stesting.Reactor{},
			ExpectedSecret: createEncryptionSecret(t, "cluster-2", make(map[string]string)),
			ExpectedError:  nil,
		},

		// First cluster: Success case when there are no cluster secrets
		// present but new cluster is about to be created
		{
			CustomObject:   createCustomObject("cluster-1"),
			PresentSecrets: []*v1.Secret{},
			APIReactors:    []k8stesting.Reactor{},
			ExpectedSecret: nil,
			ExpectedError:  nil,
		},

		// New cluster: Success case when there are three cluster secrets
		// present but expected secret doesn't exist (yet)
		{
			CustomObject: createCustomObject("cluster-4"),
			PresentSecrets: []*v1.Secret{
				createEncryptionSecret(t, "cluster-1", make(map[string]string)),
				createEncryptionSecret(t, "cluster-2", make(map[string]string)),
				createEncryptionSecret(t, "cluster-3", make(map[string]string)),
			},
			APIReactors:    []k8stesting.Reactor{},
			ExpectedSecret: nil,
			ExpectedError:  nil,
		},

		// API Error: Kubernetes API client returns unknown error
		{
			CustomObject:   createCustomObject("cluster-4"),
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

	for i, tc := range testCases {
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
			t.Errorf("TestCase %d: Resource construction failed: %#v", (i + 1), err)
			continue
		}

		state, err := r.GetCurrentState(context.TODO(), tc.CustomObject)
		if microerror.Cause(err) != tc.ExpectedError {
			t.Errorf("TestCase %d: GetCurrentState() returned error %#v - expected: %#v", (i + 1), err, tc.ExpectedError)
			continue
		}

		if state == nil && tc.ExpectedSecret == nil {
			continue
		}

		secret, ok := state.(*v1.Secret)
		if !ok {
			t.Errorf("TestCase %d: GetCurrentState() returned wrong type %T for current state. Expected %T", (i + 1), state, secret)
			continue
		}

		if tc.ExpectedSecret.Labels[randomkeytpr.ClusterIDLabel] != secret.Labels[randomkeytpr.ClusterIDLabel] {
			t.Errorf("TestCase %d: Expected secret with cluster ID label %s, found %s",
				(i + 1), tc.ExpectedSecret.Labels[randomkeytpr.ClusterIDLabel],
				secret.Labels[randomkeytpr.ClusterIDLabel],
			)
		}
	}
}

func createCustomObject(clusterID string) *v1alpha1.KVMClusterConfig {
	return &v1alpha1.KVMClusterConfig{
		Spec: v1alpha1.KVMClusterConfigSpec{
			Guest: v1alpha1.KVMClusterConfigSpecGuest{
				ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
					ID: clusterID,
				},
			},
		},
	}
}

func createEncryptionSecret(t *testing.T, clusterID string, data map[string]string) *v1.Secret {
	t.Helper()
	return createSecret(t, createCustomObject(clusterID), map[string]string{
		randomkeytpr.ClusterIDLabel: clusterID,
		randomkeytpr.KeyLabel:       randomkeytpr.EncryptionKey.String(),
	}, data)
}

func createSecret(t *testing.T, customObject *v1alpha1.KVMClusterConfig, labels, data map[string]string) *v1.Secret {
	t.Helper()
	return &v1.Secret{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "secret",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      key.EncryptionKeySecretName(*customObject),
			Namespace: v1.NamespaceDefault,
			Labels:    labels,
		},
		StringData: data,
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
