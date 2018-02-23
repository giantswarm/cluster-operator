package certconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/runtime"
	clientgofake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func Test_GetCurrentState(t *testing.T) {
	testCases := []struct {
		description         string
		clusterGuestConfig  *v1alpha1.ClusterGuestConfig
		presentCertConfigs  []*v1alpha1.CertConfig
		apiReactors         []k8stesting.Reactor
		expectedCertConfigs []*v1alpha1.CertConfig
		expectedError       error
	}{
		{
			description: "return correct certconfigs from group of certconfigs for different clusters",
			clusterGuestConfig: &v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			presentCertConfigs: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.EtcdCert),
				newCertConfig("cluster-2", certs.APICert),
				newCertConfig("cluster-2", certs.EtcdCert),
			},
			apiReactors: []k8stesting.Reactor{},
			expectedCertConfigs: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.EtcdCert),
			},
			expectedError: nil,
		},
		{
			description: "return empty list as state when there are no certconfigs present",
			clusterGuestConfig: &v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			presentCertConfigs:  []*v1alpha1.CertConfig{},
			apiReactors:         []k8stesting.Reactor{},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			expectedError:       unknownAPIError,
		},
		{
			description: "return all certconfigs that match clusterID despite of having uknonwn Cert name",
			clusterGuestConfig: &v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			presentCertConfigs: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.EtcdCert),
				newCertConfig("cluster-1", "uknown"),
				newCertConfig("cluster-2", certs.APICert),
				newCertConfig("cluster-2", certs.EtcdCert),
			},
			apiReactors: []k8stesting.Reactor{},
			expectedCertConfigs: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.EtcdCert),
				newCertConfig("cluster-1", "uknown"),
			},
			expectedError: nil,
		},
		{
			description: "handle unknown error from k8s API",
			clusterGuestConfig: &v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			presentCertConfigs:  []*v1alpha1.CertConfig{},
			apiReactors:         []k8stesting.Reactor{alwaysReturnErrorReactor(unknownAPIError)},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			expectedError:       unknownAPIError,
		},
	}

	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tc.presentCertConfigs))
			for _, cc := range tc.presentCertConfigs {
				objs = append(objs, cc)
			}

			client := fake.NewSimpleClientset(objs...)
			client.ReactionChain = append(tc.apiReactors, client.ReactionChain...)

			r, err := New(Config{
				G8sClient:   client,
				K8sClient:   clientgofake.NewSimpleClientset(),
				Logger:      logger,
				ProjectName: "cluster-operator",
				ToClusterGuestConfigFunc: func(v interface{}) (*v1alpha1.ClusterGuestConfig, error) {
					return v.(*v1alpha1.ClusterGuestConfig), nil
				},
			})

			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}

			state, err := r.GetCurrentState(context.TODO(), tc.clusterGuestConfig)
			if microerror.Cause(err) != tc.expectedError {
				t.Fatalf("GetCurrentState() returned error %#v - expected: %#v", err, tc.expectedError)
			}

			if tc.expectedError != nil && state != nil {
				t.Fatalf("GetCurrentState() must return nil state when error is returned")
			} else if tc.expectedError != nil && state == nil {
				return
			}

			certConfigs, ok := state.([]*v1alpha1.CertConfig)
			if !ok {
				t.Fatal("state type doesn't match expected: []*v1alpha1.CertConfig")
			}

			// Order doesn't matter. Important is that exactly the expected
			// ones are returned.
			for _, cc := range certConfigs {
				found := false
				for j := 0; j < len(tc.expectedCertConfigs); j++ {
					if reflect.DeepEqual(cc, tc.expectedCertConfigs[j]) {
						tc.expectedCertConfigs = append(tc.expectedCertConfigs[:j], tc.expectedCertConfigs[j+1:]...)
						found = true
						break
					}
				}

				// Extraneous CertConfig that should not exist in returned state.
				if !found {
					t.Errorf("unexpected CertConfig among return values: %#v", cc)
				}
			}

			// Missing expected CertConfigs.
			if len(tc.expectedCertConfigs) != 0 {
				for _, cc := range tc.expectedCertConfigs {
					t.Errorf("returned state doesn't have expected CertConfig: %#v", cc)
				}
			}
		})
	}
}
