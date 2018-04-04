package encryptionkey

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/resource/v1/key"
)

const (
	// AESCBCKeyLength represent the 32bytes length for AES-CBC with PKCS#7
	// padding encryption key.
	AESCBCKeyLength = 32
)

// GetDesiredState takes observed (during create, delete and update events)
// custom object as an input and returns computed desired state for it.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing desired encryption key secret")
	secretName := key.EncryptionKeySecretName(clusterGuestConfig)

	keyBytes, err := newRandomKey(AESCBCKeyLength)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secret := &v1.Secret{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      secretName,
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				label.Cluster:          key.ClusterID(clusterGuestConfig),
				label.LegacyClusterID:  key.ClusterID(clusterGuestConfig),
				label.LegacyClusterKey: label.RandomKeyTypeEncryption,
				label.ManagedBy:        r.projectName,
				label.RandomKey:        label.RandomKeyTypeEncryption,
			},
		},
		StringData: map[string]string{
			label.RandomKeyTypeEncryption: keyBytes,
		},
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computed desired encryption key secret")

	return secret, nil
}

func newRandomKey(length int) (string, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return base64.StdEncoding.EncodeToString([]byte(key)), nil
}
