package certconfig

import (
	"errors"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
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
	return newCertConfigWithVersion(clusterID, cert, "1.0.0")
}

func newCertConfigWithVersion(clusterID string, cert certs.Cert, version string) *v1alpha1.CertConfig {
	clusterGuestConfig := v1alpha1.ClusterGuestConfig{
		ID: clusterID,
	}

	labels := map[string]string{
		// Legacy
		label.LegacyClusterID: clusterID,

		// Current
		label.ClusterID: clusterID,
		label.ManagedBy: "cluster-operator",
	}

	return &v1alpha1.CertConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.CertConfigName(key.ClusterID(clusterGuestConfig), cert),
			Namespace: v1.NamespaceDefault,
			Labels:    labels,
		},
		Spec: v1alpha1.CertConfigSpec{
			Cert: v1alpha1.CertConfigSpecCert{
				ClusterID: clusterID,
			},
			VersionBundle: v1alpha1.CertConfigSpecVersionBundle{
				Version: version,
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

func verifyCertConfigCreatedReactor(t *testing.T, certConfigSeen map[string]bool) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "create",
		Resource: "certconfigs",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			createAction, ok := action.(k8stesting.CreateActionImpl)
			if !ok {
				return false, nil, microerror.Maskf(wrongTypeError, "action != k8stesting.CreateActionImpl")
			}

			var createdCertConfig *v1alpha1.CertConfig
			if createActionObj := createAction.GetObject(); createActionObj != nil {
				createdCertConfig, ok = createActionObj.(*v1alpha1.CertConfig)
				if !ok {
					return false, nil, microerror.Maskf(wrongTypeError, "CreateAction did not contain *v1alpha1.CertConfig")
				}
			}

			_, exists := certConfigSeen[createdCertConfig.Name]
			if exists {
				certConfigSeen[createdCertConfig.Name] = true
			} else {
				t.Fatalf("create(certconfig) that doesn't exist in verification table; name %s", createdCertConfig.Name)
			}

			return true, createdCertConfig, nil
		},
	}
}

func verifyCertConfigDeletedReactor(t *testing.T, certConfigSeen map[string]bool) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "delete",
		Resource: "certconfigs",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			deleteAction, ok := action.(k8stesting.DeleteActionImpl)
			if !ok {
				return false, nil, microerror.Maskf(wrongTypeError, "action != k8stesting.DeleteActionImpl")
			}

			_, exists := certConfigSeen[deleteAction.Name]
			if exists {
				certConfigSeen[deleteAction.Name] = true
			} else {
				t.Fatalf("delete(certconfig) that doesn't exist in verification table; name %s", deleteAction.Name)
			}

			return true, nil, nil
		},
	}
}
