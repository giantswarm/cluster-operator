package awsclusterconfig

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/clusterclient/service/release/searcher"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/versionbundle"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		if !key.IsProviderSpecForAWS(cr) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cluster %#q is not for aws", cr.Name))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		if key.ClusterID(&cr) == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cluster %#q misses the cluster id label", cr.Name))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	_, err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cr.Namespace).Get(key.AWSClusterConfigName(cr), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		// fall through
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("aws cluster config for cluster %#q already created", cr.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating aws cluster config for cluster %#q", cr.Name))

		awsClusterConfig, err := r.newAWSClusterConfigFromCluster(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cr.Namespace).Create(awsClusterConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created aws cluster config for cluster %#q", cr.Name))
	}

	return nil
}

func (r *Resource) newAWSClusterConfigFromCluster(ctx context.Context, cr clusterv1alpha1.Cluster) (*corev1alpha1.AWSClusterConfig, error) {
	var versionBundles []versionbundle.Bundle
	{
		req := searcher.Request{
			ReleaseVersion: key.ReleaseVersion(&cr),
		}

		res, err := r.clusterClient.Release.Searcher.Search(ctx, req)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		versionBundles = res.VersionBundles
	}

	awsClusterConfig := &corev1alpha1.AWSClusterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSClusterConfig",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.AWSClusterConfigName(cr),
			Namespace: cr.Namespace,
			Labels: map[string]string{
				label.Cluster:         key.ClusterID(&cr),
				label.OperatorVersion: key.OperatorVersion(&cr),
				label.Organization:    key.OrganizationID(&cr),
				label.ReleaseVersion:  key.ReleaseVersion(&cr),
			},
		},
		Spec: corev1alpha1.AWSClusterConfigSpec{
			Guest: corev1alpha1.AWSClusterConfigSpecGuest{
				CredentialSecret: corev1alpha1.AWSClusterConfigSpecGuestCredentialSecret{
					Name:      key.ClusterCredentialSecretName(cr),
					Namespace: key.ClusterCredentialSecretNamespace(cr),
				},
				ClusterGuestConfig: corev1alpha1.ClusterGuestConfig{
					DNSZone:        key.ClusterDNSZone(cr),
					ID:             key.ClusterID(&cr),
					VersionBundles: transformVersionBundles(versionBundles),
				},
			},
			VersionBundle: corev1alpha1.AWSClusterConfigSpecVersionBundle{
				Version: key.OperatorVersion(&cr),
			},
		},
	}

	return awsClusterConfig, nil
}

func transformVersionBundles(versionBundles []versionbundle.Bundle) []corev1alpha1.ClusterGuestConfigVersionBundle {
	var list []corev1alpha1.ClusterGuestConfigVersionBundle

	for _, b := range versionBundles {
		bundle := corev1alpha1.ClusterGuestConfigVersionBundle{
			Name:    b.Name,
			Version: b.Version,
		}

		list = append(list, bundle)
	}

	return list
}
