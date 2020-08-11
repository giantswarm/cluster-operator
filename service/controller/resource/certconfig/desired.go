package certconfig

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/key"
	"github.com/giantswarm/cluster-operator/service/internal/hamaster"
	"github.com/giantswarm/cluster-operator/service/internal/releaseversion"
)

// getDesiredState returns all desired CertConfigs for managed certificates.
func (r *Resource) getDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// When the CertConfig CR is deleted we do not need to compute the desired
	// state, because we only use the current state to delete the CR. Also note
	// that the desired state relies on the releaseversions resource, because we
	// put the cert-operator version into the CR. The releaseversions resource
	// does not fill the controller context with versions on delete events, which
	// is also why we cannot compute the correct desired state. We do not want to
	// fetch the version information on delete events to reduce eventual friction.
	// Cluster deletion should not be affected only because some releases are
	// missing or broken when fetching them from cluster-service.
	if key.IsDeleted(&cr) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not computing desired state", "reason", "the current state is used for deletion")
		return nil, nil
	}

	// We need to determine if we want to generate certificates for a Tenant
	// Cluster with a HA Master setup.
	var haMasterEnabled bool
	{
		haMasterEnabled, err = r.haMaster.Enabled(ctx, key.ClusterID(&cr))
		if hamaster.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "not computing desired state", "reason", "control plane CR not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	bd, err := r.baseDomain.BaseDomain(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	componentVersions, err := r.releaseVersion.ComponentVersion(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	certOperatorVersion := componentVersions[releaseversion.CertOperator]
	var certConfigs []*corev1alpha1.CertConfig
	{
		certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForAPI(ctx, bd, cr)))
		certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForAppOperator(ctx, bd, cr)))
		certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForAWSOperator(ctx, bd, cr)))
		certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForCalico(ctx, bd, cr)))
		certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForClusterOperator(ctx, bd, cr)))
		certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForNodeOperator(ctx, bd, cr)))
		certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForPrometheus(ctx, bd, cr)))
		certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForServiceAccount(ctx, bd, cr)))
		certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForWorker(ctx, bd, cr)))

		if haMasterEnabled {
			certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForEtcd1(ctx, bd, cr)))
			certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForEtcd2(ctx, bd, cr)))
			certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForEtcd3(ctx, bd, cr)))
		} else {
			certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForEtcd(ctx, bd, cr)))
		}

		if r.provider == label.ProviderKVM {
			certConfigs = append(certConfigs, newCertConfig(certOperatorVersion, cr, r.newSpecForFlanneldEtcdClient(ctx, bd, cr)))
		}
	}

	return certConfigs, nil
}

func newCertConfig(certOperatorVersion string, cr apiv1alpha2.Cluster, cert corev1alpha1.CertConfigSpecCert) *corev1alpha1.CertConfig {
	return &corev1alpha1.CertConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CertConfig",
			APIVersion: "core.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.CertConfigName(&cr, cert.ClusterComponent),
			Namespace: cr.Namespace,
			Labels: map[string]string{
				label.Certificate:         cert.ClusterComponent,
				label.CertOperatorVersion: certOperatorVersion,
				label.Cluster:             key.ClusterID(&cr),
				label.ManagedBy:           project.Name(),
				label.Organization:        key.OrganizationID(&cr),
			},
		},
		Spec: corev1alpha1.CertConfigSpec{
			Cert: cert,
			VersionBundle: corev1alpha1.CertConfigSpecVersionBundle{
				Version: certOperatorVersion,
			},
		},
	}
}

func (r *Resource) newSpecForAPI(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	defaultAltNames := key.CertDefaultAltNames(r.clusterDomain)
	desiredAltNames := append(defaultAltNames,
		fmt.Sprintf("master.%s", key.ClusterID(&cr)),
		fmt.Sprintf("internal-api.%s.k8s.%s", key.ClusterID(&cr), bd),
	)

	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		AltNames:         desiredAltNames,
		ClusterComponent: certs.APICert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("api.%s.k8s.%s", key.ClusterID(&cr), bd),
		IPSANs:           []string{r.apiIP, key.LocalhostIP},
		Organizations:    []string{"system:masters"},
		TTL:              r.certTTL,
	}
}

func (r *Resource) newSpecForAppOperator(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.AppOperatorAPICert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("app-operator.%s.k8s.%s", key.ClusterID(&cr), bd),
		// TODO drop system:masters once RBAC rules are in place in tenant clusters.
		//
		//     https://github.com/giantswarm/giantswarm/issues/6822
		//
		Organizations: []string{"system:masters"},
		TTL:           r.certTTL,
	}
}

func (r *Resource) newSpecForAWSOperator(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.AWSOperatorAPICert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("aws-operator.%s.k8s.%s", key.ClusterID(&cr), bd),
		// TODO drop system:masters once RBAC rules are in place in tenant clusters.
		//
		//     https://github.com/giantswarm/giantswarm/issues/6822
		//
		Organizations: []string{"system:masters"},
		TTL:           r.certTTL,
	}
}

func (r *Resource) newSpecForCalico(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.CalicoEtcdClientCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("calico.%s.k8s.%s", key.ClusterID(&cr), bd),
		TTL:              r.certTTL,
	}
}

func (r *Resource) newSpecForClusterOperator(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.ClusterOperatorAPICert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("cluster-operator.%s.k8s.%s", key.ClusterID(&cr), bd),
		// TODO drop system:masters once RBAC rules are in place in tenant clusters.
		//
		//     https://github.com/giantswarm/giantswarm/issues/6822
		//
		Organizations: []string{"system:masters"},
		TTL:           r.certTTL,
	}
}

func (r *Resource) newSpecForEtcd(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.EtcdCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("etcd.%s.k8s.%s", key.ClusterID(&cr), bd),
		IPSANs:           []string{"127.0.0.1"},
		TTL:              r.certTTL,
	}
}

func (r *Resource) newSpecForEtcd1(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.Etcd1Cert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("etcd.%s.k8s.%s", key.ClusterID(&cr), bd),
		AltNames: []string{
			fmt.Sprintf("etcd1.%s.k8s.%s", key.ClusterID(&cr), bd),
		},
		IPSANs: []string{"127.0.0.1"},
		TTL:    r.certTTL,
	}
}

func (r *Resource) newSpecForEtcd2(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.Etcd2Cert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("etcd.%s.k8s.%s", key.ClusterID(&cr), bd),
		AltNames: []string{
			fmt.Sprintf("etcd2.%s.k8s.%s", key.ClusterID(&cr), bd),
		},
		IPSANs: []string{"127.0.0.1"},
		TTL:    r.certTTL,
	}
}

func (r *Resource) newSpecForEtcd3(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.Etcd3Cert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("etcd.%s.k8s.%s", key.ClusterID(&cr), bd),
		AltNames: []string{
			fmt.Sprintf("etcd3.%s.k8s.%s", key.ClusterID(&cr), bd),
		},
		IPSANs: []string{"127.0.0.1"},
		TTL:    r.certTTL,
	}
}

func (r *Resource) newSpecForFlanneldEtcdClient(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.FlanneldEtcdClientCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("flanneld-etcd-client.%s.k8s.%s", key.ClusterID(&cr), bd),
		TTL:              r.certTTL,
	}
}

func (r *Resource) newSpecForNodeOperator(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.NodeOperatorCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("node-operator.%s.k8s.%s", key.ClusterID(&cr), bd),
		// TODO drop system:masters once RBAC rules are in place in tenant clusters.
		//
		//     https://github.com/giantswarm/giantswarm/issues/6822
		//
		Organizations: []string{"system:masters"},
		TTL:           r.certTTL,
	}
}

func (r *Resource) newSpecForPrometheus(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.PrometheusCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("prometheus.%s.k8s.%s", key.ClusterID(&cr), bd),
		// TODO drop system:masters once RBAC rules are in place in tenant clusters.
		//
		//     https://github.com/giantswarm/giantswarm/issues/6822
		//
		Organizations: []string{"system:masters"},
		TTL:           r.certTTL,
	}
}

func (r *Resource) newSpecForServiceAccount(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		ClusterComponent: certs.ServiceAccountCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("service-actxount.%s.k8s.%s", key.ClusterID(&cr), bd),
		TTL:              r.certTTL,
	}
}

func (r *Resource) newSpecForWorker(ctx context.Context, bd string, cr apiv1alpha2.Cluster) corev1alpha1.CertConfigSpecCert {
	return corev1alpha1.CertConfigSpecCert{
		AllowBareDomains: true,
		AltNames:         key.CertDefaultAltNames(r.clusterDomain),
		ClusterComponent: certs.WorkerCert.String(),
		ClusterID:        key.ClusterID(&cr),
		CommonName:       fmt.Sprintf("worker.%s.k8s.%s", key.ClusterID(&cr), bd),
		TTL:              r.certTTL,
	}
}
