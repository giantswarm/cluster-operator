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
	clientgofake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/giantswarm/cluster-operator/pkg/resource/v1/key"
)

func Test_ApplyCreateChange_Creates_createChange(t *testing.T) {
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	clusterGuestConfig := v1alpha1.ClusterGuestConfig{
		ID: "cluster-1",
	}

	createChange := []*v1alpha1.CertConfig{
		newCertConfig("cluster-1", certs.APICert),
		newCertConfig("cluster-1", certs.EtcdCert),
		newCertConfig("cluster-1", certs.PrometheusCert),
		newCertConfig("cluster-1", certs.WorkerCert),
	}

	verificationTable := map[string]bool{
		key.CertConfigName(key.ClusterID(clusterGuestConfig), certs.APICert):        false,
		key.CertConfigName(key.ClusterID(clusterGuestConfig), certs.EtcdCert):       false,
		key.CertConfigName(key.ClusterID(clusterGuestConfig), certs.PrometheusCert): false,
		key.CertConfigName(key.ClusterID(clusterGuestConfig), certs.WorkerCert):     false,
	}

	client := fake.NewSimpleClientset()
	client.ReactionChain = append([]k8stesting.Reactor{
		verifyCertConfigCreatedReactor(t, verificationTable),
	}, client.ReactionChain...)

	r, err := New(Config{
		BaseClusterConfig: newClusterConfig(),
		G8sClient:         client,
		K8sClient:         clientgofake.NewSimpleClientset(),
		Logger:            logger,
		ProjectName:       "cluster-operator",
		ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
			return v.(v1alpha1.ClusterGuestConfig), nil
		},
	})

	if err != nil {
		t.Fatalf("Resource construction failed: %#v", err)
	}

	err = r.ApplyCreateChange(context.TODO(), clusterGuestConfig, createChange)
	if err != nil {
		t.Fatalf("ApplyCreateChange(...) == %#v, want nil", err)
	}

	for k, v := range verificationTable {
		// Was CoreV1alpha1().CertConfigs(...).Create() called for given
		// CertConfig?
		if !v {
			t.Fatalf("ApplyCreateChange(...) didn't create CertConfig(%s)", k)
		}
	}
}

func Test_ApplyCreateChange_Does_Not_Make_API_Call_With_Empty_CreateChange(t *testing.T) {
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	clusterGuestConfig := v1alpha1.ClusterGuestConfig{
		ID: "cluster-1",
	}

	createChange := []*v1alpha1.CertConfig{}

	client := fake.NewSimpleClientset()
	client.ReactionChain = append([]k8stesting.Reactor{
		alwaysReturnErrorReactor(unknownAPIError),
	}, client.ReactionChain...)

	r, err := New(Config{
		BaseClusterConfig: newClusterConfig(),
		G8sClient:         client,
		K8sClient:         clientgofake.NewSimpleClientset(),
		Logger:            logger,
		ProjectName:       "cluster-operator",
		ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
			return v.(v1alpha1.ClusterGuestConfig), nil
		},
	})

	if err != nil {
		t.Fatalf("Resource construction failed: %#v", err)
	}

	err = r.ApplyCreateChange(context.TODO(), clusterGuestConfig, createChange)
	if err != nil {
		t.Fatalf("ApplyCreateChange(...) == %#v, want nil", err)
	}
}

func Test_ApplyCreateChange_Handles_K8S_API_Error(t *testing.T) {
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	clusterGuestConfig := v1alpha1.ClusterGuestConfig{
		ID: "cluster-1",
	}

	createChange := []*v1alpha1.CertConfig{
		newCertConfig("cluster-1", certs.APICert),
	}

	client := fake.NewSimpleClientset()
	client.ReactionChain = append([]k8stesting.Reactor{
		alwaysReturnErrorReactor(unknownAPIError),
	}, client.ReactionChain...)

	r, err := New(Config{
		BaseClusterConfig: newClusterConfig(),
		G8sClient:         client,
		K8sClient:         clientgofake.NewSimpleClientset(),
		Logger:            logger,
		ProjectName:       "cluster-operator",
		ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
			return v.(v1alpha1.ClusterGuestConfig), nil
		},
	})

	if err != nil {
		t.Fatalf("Resource construction failed: %#v", err)
	}

	err = r.ApplyCreateChange(context.TODO(), clusterGuestConfig, createChange)
	if microerror.Cause(err) != unknownAPIError {
		t.Fatalf("ApplyCreateChange(...) == %#v, want %#v", err, unknownAPIError)
	}
}

func Test_newCreateChange(t *testing.T) {
	testCases := []struct {
		name                string
		clusterGuestConfig  v1alpha1.ClusterGuestConfig
		currentState        interface{}
		desiredState        interface{}
		expectedCertConfigs []*v1alpha1.CertConfig
		errorMatcher        func(error) bool
	}{
		{
			name: "case 0: No certconfigs exist, single certconfig desired",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: nil,
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: One certconfig exists and it's the desired one",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			errorMatcher:        nil,
		},
		{
			name: "case 2: Some of desired certconfigs exist",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
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
				newCertConfig("cluster-1", certs.FlanneldCert),
				newCertConfig("cluster-1", certs.NodeOperatorCert),
				newCertConfig("cluster-1", certs.PrometheusCert),
				newCertConfig("cluster-1", certs.ServiceAccountCert),
				newCertConfig("cluster-1", certs.WorkerCert),
			},
			errorMatcher: nil,
		},
		{
			name: "case 3: desiredState is wrong type",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
			},
			desiredState: []string{
				"foo",
				"bar",
				"baz",
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			errorMatcher:        IsWrongType,
		},
		{
			name: "case 4: currentState is wrong type",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
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

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(Config{
				BaseClusterConfig: newClusterConfig(),
				G8sClient:         fake.NewSimpleClientset(),
				K8sClient:         clientgofake.NewSimpleClientset(),
				Logger:            logger,
				ProjectName:       "cluster-operator",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
			})

			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}

			certConfigs, err := r.newCreateChange(context.TODO(), tt.clusterGuestConfig, tt.currentState, tt.desiredState)

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

			for _, c := range tt.expectedCertConfigs {
				found := false
				for i := 0; i < len(certConfigs); i++ {
					if reflect.DeepEqual(certConfigs[i], c) {
						certConfigs = append(certConfigs[:i], certConfigs[i+1:]...)
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("%#v not found in certConfigs returned by newCreateChange", c)
				}
			}

			if len(certConfigs) > 0 {
				for _, c := range certConfigs {
					t.Errorf("unwanted certconfig present: %#v", c)
				}
			}
		})
	}
}
