package certconfig

import (
	"context"
	"fmt"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/key"
)

// GetDesiredState returns all desired CertConfigs for managed certificates.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// When the CertConfig CR is deleted we do not need to compute the desired
	// state, because we only use the current state to delete the CR. Also note
	// that the desired state relies on the operatorversions resource, because we
	// put the cert-operator version into the CR. The operatorversions resource
	// does not fill the controller context with versions on delete events, which
	// is also why we cannot compute the correct desired state. We do not want to
	// fetch the version information on delete events to reduce eventual friction.
	// Cluster deletion should not be affected only because some releases are
	// missing or broken when fetching them from cluster-service.
	if key.IsDeleted(&cr) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not computing desired state of cert config crs due to delete event")
		return nil, nil
	}

	var certConfigs []*g8sv1alpha1.CertConfig
	{
		certConfigs = append(certConfigs, newCertConfig(*cc, cr, r.newSpecForAPI(cr)))
		certConfigs = append(certConfigs, newCertConfig(*cc, cr, r.newSpecForAppOperator(cr)))
		certConfigs = append(certConfigs, newCertConfig(*cc, cr, r.newSpecForCalico(cr)))
		certConfigs = append(certConfigs, newCertConfig(*cc, cr, r.newSpecForClusterOperator(cr)))
		certConfigs = append(certConfigs, newCertConfig(*cc, cr, r.newSpecForEtcd(cr)))
		certConfigs = append(certConfigs, newCertConfig(*cc, cr, r.newSpecForNodeOperator(cr)))
		certConfigs = append(certConfigs, newCertConfig(*cc, cr, r.newSpecForPrometheus(cr)))
		certConfigs = append(certConfigs, newCertConfig(*cc, cr, r.newSpecForServiceAccount(cr)))
		certConfigs = append(certConfigs, newCertConfig(*cc, cr, r.newSpecForWorker(cr)))
	}

	if r.provider == label.ProviderKVM {
		certConfigs = append(certConfigs, newCertConfig(*cc, cr, r.newSpecForFlanneldEtcdClient(cr)))
	}

	return certConfigs, nil
}

func newCertConfig(cc controllercontext.Context, cr cmav1alpha1.Cluster, cert g8sv1alpha1.CertConfigSpecCert) *g8sv1alpha1.CertConfig {
	return &g8sv1alpha1.CertConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       "CertConfig",
			APIVersion: "core.giantswarm.io",
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name:      key.CertConfigName(&cr, cert.ClusterComponent),
			Namespace: cr.Namespace,
			Labels: map[string]string{
				label.Certificate:  cert.ClusterComponent,
				label.Cluster:      key.ClusterID(&cr),
				label.ManagedBy:    project.Name(),
				label.Organization: key.OrganizationID(&cr),
			},
		},
		Spec: g8sv1alpha1.CertConfigSpec{
			Cert: cert,
			VersionBundle: g8sv1alpha1.CertConfigSpecVersionBundle{
				Version: cc.Status.Versions["cert-operator.giantswarm.io/version"],
			},
		},
	}
}

func (r *Resource) newSpecForAPI(cr cmav1alpha1.Cluster) g8sv1alpha1.CertConfigSpecCert {
	return g8sv1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		AltNames:         key.CertAltNames(fmt.Sprintf("master.%s", key.ClusterID(&cr)), fmt.Sprintf("internal-api.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr))),
		ClusterComponent: certs.APICert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("api.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr)),
		IPSANs:           []string{r.apiIP, key.LocalhostIP},
		Organizations:    []string{"system:masters"},
		TTL:              r.certTTL,
	}
}

func (r *Resource) newSpecForAppOperator(cr cmav1alpha1.Cluster) g8sv1alpha1.CertConfigSpecCert {
	return g8sv1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.AppOperatorAPICert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("app-operator.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr)),
		// TODO drop system:masters once RBAC rules are in place in tenant clusters.
		//
		//     https://github.com/giantswarm/giantswarm/issues/6822
		//
		Organizations: []string{"system:masters"},
		TTL:           r.certTTL,
	}
}

func (r *Resource) newSpecForCalico(cr cmav1alpha1.Cluster) g8sv1alpha1.CertConfigSpecCert {
	return g8sv1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.CalicoEtcdClientCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("calico.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr)),
		TTL:              r.certTTL,
	}
}

func (r *Resource) newSpecForClusterOperator(cr cmav1alpha1.Cluster) g8sv1alpha1.CertConfigSpecCert {
	return g8sv1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.ClusterOperatorAPICert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("cluster-operator.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr)),
		// TODO drop system:masters once RBAC rules are in place in tenant clusters.
		//
		//     https://github.com/giantswarm/giantswarm/issues/6822
		//
		Organizations: []string{"system:masters"},
		TTL:           r.certTTL,
	}
}

func (r *Resource) newSpecForEtcd(cr cmav1alpha1.Cluster) g8sv1alpha1.CertConfigSpecCert {
	return g8sv1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.EtcdCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("etcd.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr)),
		IPSANs:           []string{"127.0.0.1"},
		TTL:              r.certTTL,
	}
}

func (r *Resource) newSpecForFlanneldEtcdClient(cr cmav1alpha1.Cluster) g8sv1alpha1.CertConfigSpecCert {
	return g8sv1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.FlanneldEtcdClientCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("flanneld-etcd-client.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr)),
		TTL:              r.certTTL,
	}
}

func (r *Resource) newSpecForNodeOperator(cr cmav1alpha1.Cluster) g8sv1alpha1.CertConfigSpecCert {
	return g8sv1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.NodeOperatorCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("node-operator.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr)),
		// TODO drop system:masters once RBAC rules are in place in tenant clusters.
		//
		//     https://github.com/giantswarm/giantswarm/issues/6822
		//
		Organizations: []string{"system:masters"},
		TTL:           r.certTTL,
	}
}

func (r *Resource) newSpecForPrometheus(cr cmav1alpha1.Cluster) g8sv1alpha1.CertConfigSpecCert {
	return g8sv1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.PrometheusCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("prometheus.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr)),
		// TODO drop system:masters once RBAC rules are in place in tenant clusters.
		//
		//     https://github.com/giantswarm/giantswarm/issues/6822
		//
		Organizations: []string{"system:masters"},
		TTL:           r.certTTL,
	}
}

func (r *Resource) newSpecForServiceAccount(cr cmav1alpha1.Cluster) g8sv1alpha1.CertConfigSpecCert {
	return g8sv1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.ServiceAccountCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("service-account.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr)),
		TTL:              r.certTTL,
	}
}

func (r *Resource) newSpecForWorker(cr cmav1alpha1.Cluster) g8sv1alpha1.CertConfigSpecCert {
	return g8sv1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		AltNames:         key.CertAltNames(),
		ClusterComponent: certs.WorkerCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("worker.%s.k8s.%s", key.ClusterID(&cr), key.ClusterBaseDomain(cr)),
		TTL:              r.certTTL,
	}
}
