package encryptionkey

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/key"
)

const (
	// AESCBCKeyLength represents the 32 bytes length for AES-CBC with PKCS#7
	// padding encryption key.
	AESCBCKeyLength = 32
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var secret *corev1.Secret
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computing secret %#q", secretName(cr)))

		keyBytes, err := newRandomKey(AESCBCKeyLength)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		secret = &corev1.Secret{
			ObjectMeta: apismetav1.ObjectMeta{
				Name:      secretName(cr),
				Namespace: cr.Namespace,
				Labels: map[string]string{
					label.Cluster:   key.ClusterID(&cr),
					label.ManagedBy: project.Name(),
					label.RandomKey: label.RandomKeyTypeEncryption,
				},
			},
			StringData: map[string]string{
				label.RandomKeyTypeEncryption: keyBytes,
			},
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed secret %#q", secretName(cr)))
	}

	return []*corev1.Secret{secret}, nil
}

func newRandomKey(length int) (string, error) {
	key := make([]byte, length)

	_, err := rand.Read(key)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return base64.StdEncoding.EncodeToString([]byte(key)), nil
}
