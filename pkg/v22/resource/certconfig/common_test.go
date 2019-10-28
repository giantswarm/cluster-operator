package certconfig

import (
	"errors"
	"net"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v22/key"
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
		label.LegacyComponent: string(cert),

		// Current
		label.Cluster:      clusterID,
		label.ManagedBy:    "cluster-operator",
		label.Organization: "ACME Inc.",
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

func newClusterConfig() cluster.Config {
	return cluster.Config{
		CertTTL: "",
		IP: cluster.IP{
			Range: net.IPv4(172, 31, 0, 0),
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

func verifyCertConfigUpdatedReactor(t *testing.T, certConfigSeen map[string]bool) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "update",
		Resource: "certconfigs",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			updateAction, ok := action.(k8stesting.UpdateActionImpl)
			if !ok {
				return false, nil, microerror.Maskf(wrongTypeError, "action != k8stesting.UpdateActionImpl")
			}

			var updatedCertConfig *v1alpha1.CertConfig
			if updateActionObj := updateAction.GetObject(); updateActionObj != nil {
				updatedCertConfig, ok = updateActionObj.(*v1alpha1.CertConfig)
				if !ok {
					return false, nil, microerror.Maskf(wrongTypeError, "UpdateAction did not contain *v1alpha1.CertConfig")
				}
			}

			_, exists := certConfigSeen[updatedCertConfig.Name]
			if exists {
				certConfigSeen[updatedCertConfig.Name] = true
			} else {
				t.Fatalf("update(certconfig) that doesn't exist in verification table; name %s", updatedCertConfig.Name)
			}

			return true, updatedCertConfig, nil
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
