package certconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/micrologger"
	clientgofake "k8s.io/client-go/kubernetes/fake"
)

func Test_newDeleteChange_Deletes_Existing_CertConfigs(t *testing.T) {
	testCases := []struct {
		name                string
		clusterGuestConfig  *v1alpha1.ClusterGuestConfig
		currentState        interface{}
		desiredState        interface{}
		expectedCertConfigs []*v1alpha1.CertConfig
		errorMatcher        func(error) bool
	}{
		{
			name: "case 0: No certconfigs exist, single certconfig desired",
			clusterGuestConfig: &v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: nil,
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			errorMatcher:        nil,
		},
		{
			name: "case 1: One certconfig exists and it's the same as desired one",
			clusterGuestConfig: &v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			errorMatcher: nil,
		},
		{
			name: "case 2: Some of desired certconfigs but not all",
			clusterGuestConfig: &v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
			},
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
				newCertConfig("cluster-1", certs.FlanneldCert),
				newCertConfig("cluster-1", certs.NodeOperatorCert),
				newCertConfig("cluster-1", certs.PrometheusCert),
				newCertConfig("cluster-1", certs.ServiceAccountCert),
				newCertConfig("cluster-1", certs.WorkerCert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
			},
			errorMatcher: nil,
		},
		{
			name: "case 3: currentState is wrong type",
			clusterGuestConfig: &v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []string{
				"foo",
				"bar",
				"baz",
			},
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			errorMatcher:        IsWrongType,
		},
	}

	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(Config{
				G8sClient:   fake.NewSimpleClientset(),
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

			certConfigs, err := r.newDeleteChange(context.TODO(), tt.clusterGuestConfig, tt.currentState, tt.desiredState)

			switch {
			case err == nil && tt.errorMatcher == nil: // correct; carry on
			case err != nil && tt.errorMatcher != nil:
				if !tt.errorMatcher(err) {
					t.Fatalf("error == %#v, want matching", err)
				}
			case err != nil && tt.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tt.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			}

			// Verify that certconfigs that are expected to be deleted, are the
			// only ones in the returned list of certconfigs that are to be
			// updated.  Order doesn't matter here.
			for _, c := range tt.expectedCertConfigs {
				found := false
				for i := 0; i < len(certConfigs); i++ {
					if reflect.DeepEqual(certConfigs[i], c) {
						// When matching certconfig is found, remove from list
						// returned by newDeleteChange(). When all expected
						// certconfigs are iterated, returned list must be
						// empty.
						certConfigs = append(certConfigs[:i], certConfigs[i+1:]...)
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("%#v not found in certConfigs returned by newDeleteChange", c)
				}
			}

			// Verify that there aren't any unexpected extra certconfigs going
			// to be deleted.
			if len(certConfigs) > 0 {
				for _, c := range certConfigs {
					t.Errorf("unwanted certconfig present: %#v", c)
				}
			}
		})
	}
}
