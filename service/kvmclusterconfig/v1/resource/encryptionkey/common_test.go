package encryptionkey

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/cluster-operator/service/kvmclusterconfig/v1/key"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newCustomObject(clusterID string) *v1alpha1.KVMClusterConfig {
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

func newEncryptionSecret(t *testing.T, clusterID string, data map[string]string) *v1.Secret {
	t.Helper()
	return newSecret(t, newCustomObject(clusterID), map[string]string{
		randomkeytpr.ClusterIDLabel: clusterID,
		randomkeytpr.KeyLabel:       randomkeytpr.EncryptionKey.String(),
	}, data)
}

func newSecret(t *testing.T, customObject *v1alpha1.KVMClusterConfig, labels, data map[string]string) *v1.Secret {
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
