package encryptionkey

import (
	"errors"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v16/key"
)

var (
	// Empty value pointer to v1.Secret for value check testing.
	emptySecretPointer *v1.Secret

	// Forbidden Kubernetes API call
	forbiddenAPICall = errors.New("Forbidden API call")

	// Error to return when simulating unknown error returned from Kubernetes
	// API client.
	unknownAPIError = errors.New("Unknown error from k8s API")
)

type apiReactorFactory func(t *testing.T) k8stesting.Reactor

func newClusterGuestConfig(clusterID string) v1alpha1.ClusterGuestConfig {
	return v1alpha1.ClusterGuestConfig{
		ID: clusterID,
	}
}

func newEncryptionSecret(t *testing.T, clusterID string, data map[string]string) *v1.Secret {
	t.Helper()
	return newSecret(t, newClusterGuestConfig(clusterID), map[string]string{
		label.LegacyClusterID:  clusterID,
		label.LegacyClusterKey: label.RandomKeyTypeEncryption,
	}, data)
}

func newSecret(t *testing.T, clusterGuestConfig v1alpha1.ClusterGuestConfig, labels, data map[string]string) *v1.Secret {
	t.Helper()
	return &v1.Secret{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "secret",
			APIVersion: "v1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      key.EncryptionKeySecretName(clusterGuestConfig),
			Namespace: v1.NamespaceDefault,
			Labels:    labels,
		},
		StringData: data,
	}
}

func assertSecret(t *testing.T, computedSecret, expectedSecret *v1.Secret) {
	t.Helper()

	if expectedSecret == nil && computedSecret != nil {
		t.Errorf("expected nil secret. Received %#v", computedSecret)
		return
	} else if expectedSecret != nil && computedSecret == nil {
		t.Error("expected non-nil secret. Received nil.")
		return
	}

	if !reflect.DeepEqual(computedSecret, expectedSecret) {
		t.Errorf("Computed secret %#v doesn't match expected: %#v",
			computedSecret, expectedSecret)
	}
}

func verifySecretCreatedReactor(t *testing.T, v *v1.Secret) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "create",
		Resource: "secrets",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			createAction, ok := action.(k8stesting.CreateActionImpl)
			if !ok {
				return false, nil, microerror.Maskf(wrongTypeError, "action != k8stesting.CreateActionImpl")
			}

			createdSecret, err := toSecret(createAction.GetObject())
			if err != nil {
				return false, nil, microerror.Maskf(wrongTypeError, "CreateAction did not contain *v1.Secret")
			}

			assertSecret(t, createdSecret, v)

			return true, createdSecret, nil
		},
	}
}

func verifySecretDeletedReactor(t *testing.T, v *v1.Secret) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "delete",
		Resource: "secrets",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			deleteAction, ok := action.(k8stesting.DeleteActionImpl)
			if !ok {
				return false, nil, microerror.Maskf(wrongTypeError, "action != k8stesting.DeleteActionImpl")
			}

			if v.Name != deleteAction.GetName() {
				t.Errorf("Deleted secret name '%s' doesn't match expected '%s'", deleteAction.GetName(), v.Name)
			}

			return true, nil, nil
		},
	}
}

func alwaysReturnErrorReactor(err error) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "*",
		Resource: "*",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, err
		},
	}
}
