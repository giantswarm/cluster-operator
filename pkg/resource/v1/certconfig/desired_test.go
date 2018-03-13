package certconfig

import (
	"context"
	"net"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger"
	clientgofake "k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/resource/v1/certconfig/key"
)

func Test_GetDesiredState_Returns_CertConfig_For_All_Managed_Certs(t *testing.T) {
	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	clusterGuestConfig := v1alpha1.ClusterGuestConfig{
		ID: "cluster-1",
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
		BaseClusterConfig: &cluster.Config{
			ClusterID: "cluster-1",
			CertTTL:   "720h",
			IP: cluster.IP{
				Range: clusterCIDR,
			},
		},
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

	desiredState, err := r.GetDesiredState(context.TODO(), &clusterGuestConfig)
	if err != nil {
		t.Fatalf("GetDesiredState() == %#v, want nil error", err)
	}

	certConfigs, ok := desiredState.([]*v1alpha1.CertConfig)
	if !ok {
		t.Fatalf("GetDesiredState() == %#v, wrong type %T, want %T", desiredState, desiredState, certConfigs)
	}

	for _, mc := range managedCertificates {
		certConfigName := key.CertConfigName(key.ClusterID(clusterGuestConfig), mc.name)
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
