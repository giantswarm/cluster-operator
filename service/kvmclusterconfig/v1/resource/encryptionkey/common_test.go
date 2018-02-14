package encryptionkey

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/cluster-operator/service/kvmclusterconfig/v1/key"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func assertPatch(t *testing.T, testNum int, computedPatch, expectedPatch *framework.Patch) {
	t.Helper()

	if expectedPatch == nil && computedPatch != nil {
		t.Errorf("TestCase %d: Expected nil patch. Received %#v", testNum, computedPatch)
		return
	} else if expectedPatch != nil && computedPatch == nil {
		t.Errorf("TestCase %d: Expected non-nil patch. Received nil.", testNum)
		return
	}

	if !reflect.DeepEqual(computedPatch, expectedPatch) {
		t.Errorf("TestCase %d: Computed patch %#v doesn't match expected: %#v",
			testNum, computedPatch, expectedPatch)
	}
}
