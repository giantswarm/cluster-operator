package certconfig

import (
	"context"
	"net"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/micrologger/microloggertest"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/v1/key"
)

func Test_GetDesiredState_Returns_CertConfig_For_All_Managed_Certs(t *testing.T) {
	managedCertificates := []certs.Cert{
		certs.APICert,
		certs.Cert("calico"),
		certs.ClusterOperatorAPICert,
		certs.EtcdCert,
		certs.NodeOperatorCert,
		certs.PrometheusCert,
		certs.ServiceAccountCert,
		certs.WorkerCert,
	}

	clusterGuestConfig := v1alpha1.ClusterGuestConfig{
		DNSZone: "foo.bar.example.com",
		ID:      "cluster-1",
		VersionBundles: []v1alpha1.ClusterGuestConfigVersionBundle{
			{
				Name:    certOperatorID,
				Version: "1.0.0",
			},
		},
	}

	clusterCIDR, _, err := net.ParseCIDR("172.31.0.0/16")
	if err != nil {
		t.Fatalf("failed to parse cluster CIDR: %v", err)
	}

	r, err := New(Config{
		BaseClusterConfig: cluster.Config{
			ClusterID: "cluster-1",
			CertTTL:   "720h",
			IP: cluster.IP{
				Range: clusterCIDR,
			},
		},
		G8sClient:   fake.NewSimpleClientset(),
		K8sClient:   clientgofake.NewSimpleClientset(),
		Logger:      microloggertest.New(),
		ProjectName: "cluster-operator",
		ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
			return v.(v1alpha1.ClusterGuestConfig), nil
		},
	})

	if err != nil {
		t.Fatalf("Resource construction failed: %#v", err)
	}

	desiredState, err := r.GetDesiredState(context.TODO(), clusterGuestConfig)
	if err != nil {
		t.Fatalf("GetDesiredState() == %#v, want nil error", err)
	}

	certConfigs, ok := desiredState.([]*v1alpha1.CertConfig)
	if !ok {
		t.Fatalf("GetDesiredState() == %#v, wrong type %T, want %T", desiredState, desiredState, certConfigs)
	}

	for _, cert := range managedCertificates {
		certConfigName := key.CertConfigName(key.ClusterID(clusterGuestConfig), cert)
		found := false
		for i := 0; i < len(certConfigs); i++ {
			cc := certConfigs[i]
			if cc.Name == certConfigName {
				found = true
				certConfigs = append(certConfigs[:i], certConfigs[i+1:]...)
				break
			}
		}

		if !found {
			t.Fatalf("GetDesiredState() doesn't return wanted CertConfig: %s", certConfigName)
		}
	}

	if len(certConfigs) > 0 {
		for _, cc := range certConfigs {
			t.Errorf("GetDesiredState() returns unwanted CertConfig: %#v", cc)
		}
	}
}

func Test_newServerDomain(t *testing.T) {
	testCases := []struct {
		name                 string
		inputCommonDomain    string
		inputCert            certs.Cert
		expectedServerDomain string
		errorMatcher         func(error) bool
	}{
		{
			name:                 "case 0: valid common domain foo.bar with APICert",
			inputCommonDomain:    "foo.bar",
			inputCert:            certs.APICert,
			expectedServerDomain: "api.foo.bar",
			errorMatcher:         nil,
		},
		{
			name:                 "case 1: valid common domain .bar with ServiceAccountCert",
			inputCommonDomain:    ".bar",
			inputCert:            certs.ServiceAccountCert,
			expectedServerDomain: "service-account.bar",
			errorMatcher:         nil,
		},
		{
			name:                 "case 2: valid hypothetical root domain '.' with EtcdCert",
			inputCommonDomain:    ".",
			inputCert:            certs.EtcdCert,
			expectedServerDomain: "etcd.",
			errorMatcher:         nil,
		},
		{
			name:                 "case 3: valid common domain with prefixing space ' foo.bar' with EtcdCert",
			inputCommonDomain:    " foo.bar",
			inputCert:            certs.EtcdCert,
			expectedServerDomain: "etcd.foo.bar",
			errorMatcher:         nil,
		},
		{
			name:                 "case 4: valid common domain with prefixing tab '\tfoo.bar' with EtcdCert",
			inputCommonDomain:    "\tfoo.bar",
			inputCert:            certs.EtcdCert,
			expectedServerDomain: "etcd.foo.bar",
			errorMatcher:         nil,
		},

		{
			name:                 "case 5: invalid common domain 'invalid' with EtcdCert",
			inputCommonDomain:    "invalid",
			inputCert:            certs.EtcdCert,
			expectedServerDomain: "",
			errorMatcher:         IsInvalidConfig,
		},
		{
			name:                 "case 6: empty common domain with EtcdCert",
			inputCommonDomain:    "",
			inputCert:            certs.EtcdCert,
			expectedServerDomain: "",
			errorMatcher:         IsInvalidConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serverDomain, err := newServerDomain(tc.inputCommonDomain, tc.inputCert)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if serverDomain != tc.expectedServerDomain {
				t.Fatalf("serverDomain == %q, want %q", serverDomain, tc.expectedServerDomain)
			}
		})
	}
}
