package encryptionkey

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var unknownRandError = errors.New("unknown error from crypto/rand.Read()")

func Test_GetDesiredState_Secret_Properties(t *testing.T) {
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	r, err := New(Config{
		K8sClient:   fake.NewSimpleClientset(),
		Logger:      logger,
		ProjectName: "cluster-operator",
		ToClusterGuestConfigFunc: func(v interface{}) (*v1alpha1.ClusterGuestConfig, error) {
			return v.(*v1alpha1.ClusterGuestConfig), nil
		},
	})

	if err != nil {
		t.Errorf("Resource construction failed: %#v", err)
	}

	clusterGuestConfig := &v1alpha1.ClusterGuestConfig{
		ID: "cluster-1",
	}

	testPattern := "A"
	rand.Reader = &predictableReader{
		pattern: testPattern,
	}

	state, err := r.GetDesiredState(context.TODO(), clusterGuestConfig)
	if err != nil {
		t.Fatalf("GetDesiredState() returned an error: %#v", err)
	}

	secret, ok := state.(*v1.Secret)
	if !ok {
		t.Fatalf("Unexpected state type: %T, expected %T", state, secret)
	}

	if secret.Namespace != v1.NamespaceDefault {
		t.Errorf("Secret has wrong namespace: %s, expected %s", secret.Namespace, v1.NamespaceDefault)
	}

	key, present := secret.StringData[label.RandomKeyTypeEncryption]
	if !present {
		t.Errorf("Encryption key is missing from secret")
	}

	keyData, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		t.Errorf("Encryption key decoding failed: %#v", err)
	}

	keyLen := len(keyData)
	if keyLen != AESCBCKeyLength {
		t.Errorf("Encryption key length doesn't match: %d expected %d", keyLen, AESCBCKeyLength)
	}

	for i := 0; i < AESCBCKeyLength; i++ {
		if keyData[i] != 'A' {
			t.Errorf("keyData[%d] does not match pattern %s; is crypto/rand used?", i, testPattern)
		}
	}
}

func Test_GetDesiredState_Rand_Error_Handling(t *testing.T) {
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	r, err := New(Config{
		K8sClient:   fake.NewSimpleClientset(),
		Logger:      logger,
		ProjectName: "cluster-operator",
		ToClusterGuestConfigFunc: func(v interface{}) (*v1alpha1.ClusterGuestConfig, error) {
			return v.(*v1alpha1.ClusterGuestConfig), nil
		},
	})

	if err != nil {
		t.Errorf("Resource construction failed: %#v", err)
	}

	clusterGuestConfig := &v1alpha1.ClusterGuestConfig{
		ID: "cluster-1",
	}

	// Overwrite crypto/rand.Reader in order to produce error from rand.Read()
	rand.Reader = &failingReader{}

	state, err := r.GetDesiredState(context.TODO(), clusterGuestConfig)
	if microerror.Cause(err) != unknownRandError {
		t.Errorf("Unexpected error received: %#v, expected %#v", err, unknownRandError)
	}

	if state != nil {
		t.Errorf("State returned despite of error")
	}
}

type failingReader struct{}

func (r *failingReader) Read(_ []byte) (int, error) {
	return -1, unknownRandError
}

type predictableReader struct {
	pattern string
}

func (r *predictableReader) Read(b []byte) (n int, err error) {
	if len(r.pattern) < 1 {
		panic("predictableReader.pattern cannot be empty")
	}

	for n = 0; n < len(b); n++ {
		b[n] = r.pattern[n%len(r.pattern)]
	}

	return
}
