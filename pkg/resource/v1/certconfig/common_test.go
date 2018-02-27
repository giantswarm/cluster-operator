package certconfig

import (
	"errors"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/resource/v1/certconfig/key"
)

var (
	// unknownAPIError for simulating unknown error returned from Kubernetes
	// API client.
	unknownAPIError = errors.New("Unknown error from k8s API")
)

func newCertConfig(clusterID string, cert certs.Cert) *v1alpha1.CertConfig {
	clusterGuestConfig := v1alpha1.ClusterGuestConfig{
		ID: clusterID,
	}

	labels := map[string]string{
		// Legacy
		randomkeytpr.ClusterIDLabel: clusterID,

		// Current
		label.ClusterIDLabel: clusterID,
		label.ManagedByLabel: "cluster-operator",
	}

	return &v1alpha1.CertConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.CertConfigName(clusterGuestConfig, cert),
			Namespace: v1.NamespaceDefault,
			Labels:    labels,
		},
		Spec: v1alpha1.CertConfigSpec{
			Cert: v1alpha1.CertConfigSpecCert{
				ClusterID: clusterID,
			},
		},
	}
}

func alwaysReturnErrorReactor(err error) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "*",
		Resource: "*",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, err
		},
	}
}
