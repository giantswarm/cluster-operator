package awsclusterconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v16/key"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
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

	machineDeployments, err := r.getMachineDeployments(ctx, cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	// Get existing AWSClusterConfig or create a new one.
	awsClusterConfig, err := r.getAWSClusterConfig(ctx, cluster)
	if errors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating AWSClusterConfig %q/%q", awsClusterConfig.Namespace, awsClusterConfig.Name))

		awsClusterConfig = r.constructAWSClusterConfig(cluster, machineDeployments)

		_, err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cluster.Namespace).Create(awsClusterConfig)
		if errors.IsAlreadyExists(err) {
			r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("AWSClusterConfig %q/%q already exists", awsClusterConfig.Namespace, awsClusterConfig.Name))
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created AWSClusterConfig %q/%q", awsClusterConfig.Namespace, awsClusterConfig.Name))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	// Map desired state from Cluster to AWSClusterConfig.
	r.mapClusterToAWSClusterConfig(awsClusterConfig, cluster, machineDeployments)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating AWSClusterConfig %q/%q", awsClusterConfig.Namespace, awsClusterConfig.Name))

	_, err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cluster.Namespace).Update(awsClusterConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated AWSClusterConfig %q/%q", awsClusterConfig.Namespace, awsClusterConfig.Name))

	return nil
}

// getAWSClusterConfig returns corresponding AWSClusterConfig CR for given
// Cluster if one exists or constructs new empty one.
func (r *Resource) getAWSClusterConfig(ctx context.Context, cluster clusterv1alpha1.Cluster) (*v1alpha1.AWSClusterConfig, error) {
	var awsClusterConfig *v1alpha1.AWSClusterConfig
	var err error

	awsClusterConfigName := key.AWSClusterConfigName(cluster)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if AWSClusterConfig %q/%q exists", cluster.Namespace, awsClusterConfigName))

	awsClusterConfig, err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cluster.Namespace).Get(awsClusterConfigName, v1.GetOptions{})
	if errors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find AWSClusterConfig %q/%q", cluster.Namespace, awsClusterConfigName))
		return nil, microerror.Mask(err)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found AWSClusterConfig %q/%q", cluster.Namespace, awsClusterConfigName))

	return awsClusterConfig, nil
}

func (r *Resource) getMachineDeployments(ctx context.Context, cluster clusterv1alpha1.Cluster) ([]clusterv1alpha1.MachineDeployment, error) {
	labelSelector := v1.AddLabelToSelector(&v1.LabelSelector{}, label.Cluster, key.ClusterID(cluster))
	// TODO: Add selector for provider annotation?

	listOptions := v1.ListOptions{
		LabelSelector: labelSelector.String(),
	}

	machineDeploymentList, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(cluster.Namespace).List(listOptions)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return machineDeploymentList.Items, nil
}

func (r *Resource) mapClusterToAWSClusterConfig(awsClusterConfig *v1alpha1.AWSClusterConfig, cluster clusterv1alpha1.Cluster, machineDeployments []clusterv1alpha1.MachineDeployment) {
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.AvailabilityZones = len(key.ClusterAvailabilityZones(cluster, machineDeployments))
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.DNSZone = key.ClusterDNSZone(cluster)
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.ID = key.ClusterID(cluster)
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.Name = key.ClusterName(cluster)
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.ReleaseVersion = key.ClusterReleaseVersion(cluster)

	// awsClusterConfig.Spec.Guest.ClusterGuestConfig.VersionBundles

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

func (r *Resource) constructAWSClusterConfig(cluster clusterv1alpha1.Cluster, machineDeployments []clusterv1alpha1.MachineDeployment) *v1alpha1.AWSClusterConfig {
	cc := &v1alpha1.AWSClusterConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       "AWSClusterConfig",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      key.AWSClusterConfigName(cluster),
			Namespace: cluster.Namespace,
		},
	}

	r.mapClusterToAWSClusterConfig(cc, cluster, machineDeployments)

	return cc
}
