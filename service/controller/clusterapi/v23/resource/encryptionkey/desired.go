package encryptionkey

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v23/key"
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

	// The encryptionkey resource implements a state getter which is used by a
	// generated secrets resource. This is to have a common approach of creating,
	// deleting and updating secrets. The speciality of the encryption key managed
	// in this resource here is that it must not get updated ever. So we have a
	// little hack here to return the current secret as desired secret in case it
	// already exists in Kubernetes. This prevents updates on the secret as the
	// comparison of the generic secrets resource does not find any difference
	// between current and desired state. If there is no secret in Kubernetes yet,
	// we fall through and compute the desired encryption key secret so it gets
	// created.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding secret %#q in namespace %#q", secretName(cr), cr.Namespace))

		secret, err := r.k8sClient.CoreV1().Secrets(cr.Namespace).Get(secretName(cr), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// fall through
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find secret %#q in namespace %#q", secretName(cr), cr.Namespace))
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found secret %#q in namespace %#q", secretName(cr), cr.Namespace))
			return []*corev1.Secret{secret}, nil
		}
	}

	var secret *corev1.Secret
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computing secret %#q", secretName(cr)))

		keyBytes, err := newRandomKey(AESCBCKeyLength)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName(cr),
				Namespace: cr.Namespace,
				Labels: map[string]string{
					label.Cluster:   key.ClusterID(&cr),
					label.ManagedBy: project.Name(),
					label.RandomKey: label.RandomKeyTypeEncryption,

					// TODO drop deprecated labels.
					//
					//     https://github.com/giantswarm/randomkeys/pull/15
					//
					"clusterID":  key.ClusterID(&cr),
					"clusterKey": label.RandomKeyTypeEncryption,
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
