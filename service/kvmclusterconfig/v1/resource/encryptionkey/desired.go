package encryptionkey

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/giantswarm/cluster-operator/service/kvmclusterconfig/v1/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// AESCBCKeyLength represent the 32bytes length for AES-CBC with PKCS#7
	// padding encryption key.
	AESCBCKeyLength = 32
)

// GetDesiredState takes observed (during create, delete and update events)
// custom object as an input and returns computed desired state for it.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing desired encryption key secret")
	secretName := key.EncryptionKeySecretName(customObject)

	keyBytes, err := newRandomKey(AESCBCKeyLength)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secret := &v1.Secret{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      secretName,
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				randomkeytpr.ClusterIDLabel: key.ClusterID(customObject),
				randomkeytpr.KeyLabel:       randomkeytpr.EncryptionKey.String(),
			},
		},
		StringData: map[string]string{
			randomkeytpr.EncryptionKey.String(): keyBytes,
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
