package awsclusterconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/clusterclient/service/release/searcher"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/versionbundle"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cluster, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !key.IsProviderSpecForAWS(cluster) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("provider extension in cluster cr %q is not for AWS", cluster.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	var versionBundles []versionbundle.Bundle
	{
		req := searcher.Request{
			ReleaseVersion: key.ClusterReleaseVersion(cluster),
		}

		res, err := r.clusterClient.Release.Searcher.Search(ctx, req)
		if err != nil {
			return microerror.Mask(err)
		}

		versionBundles = res.VersionBundles
	}

	// Get existing AWSClusterConfig or create a new one.
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if AWSClusterConfig %q/%q exists", cluster.Namespace, key.AWSClusterConfigName(cluster)))

	awsClusterConfig, err := r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cluster.Namespace).Get(key.AWSClusterConfigName(cluster), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find AWSClusterConfig %q/%q", cluster.Namespace, key.AWSClusterConfigName(cluster)))
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating AWSClusterConfig %q/%q", awsClusterConfig.Namespace, key.AWSClusterConfigName(cluster)))

		awsClusterConfig = r.constructAWSClusterConfig(cluster, versionBundles)

		_, err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cluster.Namespace).Create(awsClusterConfig)
		if errors.IsAlreadyExists(err) {
			r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("AWSClusterConfig %q/%q already exists", awsClusterConfig.Namespace, awsClusterConfig.Name))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created AWSClusterConfig %q/%q", awsClusterConfig.Namespace, awsClusterConfig.Name))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found AWSClusterConfig %q/%q", cluster.Namespace, awsClusterConfig.Name))

	// Map desired state from Cluster to AWSClusterConfig.
	r.mapClusterToAWSClusterConfig(awsClusterConfig, cluster, versionBundles)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating AWSClusterConfig %q/%q", awsClusterConfig.Namespace, awsClusterConfig.Name))

	_, err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cluster.Namespace).Update(awsClusterConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated AWSClusterConfig %q/%q", awsClusterConfig.Namespace, awsClusterConfig.Name))

	return nil
}

func (r *Resource) mapClusterToAWSClusterConfig(awsClusterConfig *v1alpha1.AWSClusterConfig, cluster clusterv1alpha1.Cluster, versionBundles []versionbundle.Bundle) {
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.AvailabilityZones = NumberOfAZsWithNodePools
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.DNSZone = key.ClusterDNSZone(cluster)
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.ID = key.ClusterID(cluster)
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.Name = key.ClusterName(cluster)
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.ReleaseVersion = key.ClusterReleaseVersion(cluster)

	for _, b := range versionBundles {
		bundle := v1alpha1.ClusterGuestConfigVersionBundle{
			Name:    b.Name,
			Version: b.Version,
		}
		awsClusterConfig.Spec.Guest.ClusterGuestConfig.VersionBundles = append(awsClusterConfig.Spec.Guest.ClusterGuestConfig.VersionBundles, bundle)
	}

	awsClusterConfig.Spec.Guest.CredentialSecret.Name = key.ClusterCredentialSecretName(cluster)
	awsClusterConfig.Spec.Guest.CredentialSecret.Namespace = key.ClusterCredentialSecretNamespace(cluster)

	awsClusterConfig.Spec.Guest.Masters = []v1alpha1.AWSClusterConfigSpecGuestMaster{
		{
			AWSClusterConfigSpecGuestNode: v1alpha1.AWSClusterConfigSpecGuestNode{
				InstanceType: key.ClusterMasterInstanceType(cluster),
			},
		},
	}

	// TODO: Workers shall be added when we have better understanding towards template structure.
}

func (r *Resource) constructAWSClusterConfig(cluster clusterv1alpha1.Cluster, versionBundles []versionbundle.Bundle) *v1alpha1.AWSClusterConfig {
	cc := &v1alpha1.AWSClusterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSClusterConfig",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.AWSClusterConfigName(cluster),
			Namespace: cluster.Namespace,
		},
	}

	r.mapClusterToAWSClusterConfig(cc, cluster, versionBundles)

	return cc
}
