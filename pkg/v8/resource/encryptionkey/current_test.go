package encryptionkey

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

func Test_GetCurrentState_Reads_Secrets_For_Relevant_ClusterID(t *testing.T) {
	testCases := []struct {
		description        string
		clusterGuestConfig v1alpha1.ClusterGuestConfig
		presentSecrets     []*v1.Secret
		apiReactors        []k8stesting.Reactor
		expectedSecret     *v1.Secret
		expectedError      error
	}{
		{
			description:        "three clusters exist - return secret for the one where custom object belongs",
			clusterGuestConfig: newClusterGuestConfig("cluster-2"),
			presentSecrets: []*v1.Secret{
				newEncryptionSecret(t, "cluster-1", make(map[string]string)),
				newEncryptionSecret(t, "cluster-2", make(map[string]string)),
				newEncryptionSecret(t, "cluster-3", make(map[string]string)),
			},
			apiReactors:    []k8stesting.Reactor{},
			expectedSecret: newEncryptionSecret(t, "cluster-2", make(map[string]string)),
			expectedError:  nil,
		},
		{
			description:        "no clusters exist - return empty list of secrets",
			clusterGuestConfig: newClusterGuestConfig("cluster-1"),
			presentSecrets:     []*v1.Secret{},
			apiReactors:        []k8stesting.Reactor{},
			expectedSecret:     nil,
			expectedError:      nil,
		},
		{
			description:        "three clusters exist - return secrets for them despite custom object referring to new one",
			clusterGuestConfig: newClusterGuestConfig("cluster-4"),
			presentSecrets: []*v1.Secret{
				newEncryptionSecret(t, "cluster-1", make(map[string]string)),
				newEncryptionSecret(t, "cluster-2", make(map[string]string)),
				newEncryptionSecret(t, "cluster-3", make(map[string]string)),
			},
			apiReactors:    []k8stesting.Reactor{},
			expectedSecret: nil,
			expectedError:  nil,
		},
		{
			description:        "handle unknown error returned from Kubernetes API client",
			clusterGuestConfig: newClusterGuestConfig("cluster-4"),
			presentSecrets:     []*v1.Secret{},
			apiReactors: []k8stesting.Reactor{
				alwaysReturnErrorReactor(unknownAPIError),
			},
			expectedSecret: nil,
			expectedError:  unknownAPIError,
		},
	}

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tc.presentSecrets))
			for _, s := range tc.presentSecrets {
				objs = append(objs, s)
			}

			client := fake.NewSimpleClientset(objs...)
			client.ReactionChain = append(tc.apiReactors, client.ReactionChain...)

			r, err := New(Config{
				K8sClient:   client,
				Logger:      logger,
				ProjectName: "cluster-operator",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
				ToClusterObjectMetaFunc: func(v interface{}) (apismetav1.ObjectMeta, error) {
					return apismetav1.ObjectMeta{
						Namespace: v1.NamespaceDefault,
					}, nil
				},
			})

			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}

			state, err := r.GetCurrentState(context.TODO(), tc.clusterGuestConfig)
			if microerror.Cause(err) != tc.expectedError {
				t.Fatalf("GetCurrentState() returned error %#v - expected: %#v", err, tc.expectedError)
			}

			if state == nil && tc.expectedSecret == nil {
				// Ok
				return
			}

			secret, ok := state.(*v1.Secret)
			if !ok {
				t.Fatalf("GetCurrentState() returned wrong type %T for current state. expected %T", state, secret)
			}

			if tc.expectedSecret.Labels[label.LegacyClusterID] != secret.Labels[label.LegacyClusterID] {
				t.Fatalf("expected secret with cluster ID label %s, found %s",
					tc.expectedSecret.Labels[label.LegacyClusterID],
					secret.Labels[label.LegacyClusterID],
				)
			}
		})
	}
}
