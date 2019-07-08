package awsclusterconfig

import (
	"context"
	"fmt"
	"reflect"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/clusterclient/service/release/searcher"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/versionbundle"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !key.IsProviderSpecForAWS(cr) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("provider extension in cluster %#q is not for AWS", cr.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	// ClusterID is core part of e.g. PKI initialization etc. so it must be
	// present before proceeding further.
	if key.ClusterID(&cr) == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("provider status in cluster %#q does not contain cluster ID", cr.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	var machineDeployments []clusterv1alpha1.MachineDeployment
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		l := metav1.AddLabelToSelector(
			&v1.LabelSelector{},
			label.Cluster,
			key.ClusterID(&cr),
		)
		o := metav1.ListOptions{
			LabelSelector: labels.Set(l.MatchLabels).String(),
		}

		list, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		machineDeployments = list.Items

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d MachineDeployments for tenant cluster", len(machineDeployments)))
	}

	var versionBundles []versionbundle.Bundle
	{
		req := searcher.Request{
			ReleaseVersion: key.ReleaseVersion(&cr),
		}

		res, err := r.clusterClient.Release.Searcher.Search(ctx, req)
		if err != nil {
			return microerror.Mask(err)
		}

		versionBundles = res.VersionBundles
	}

	// Get existing AWSClusterConfig or create a new one.
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if AWSClusterConfig %#q/%#q exists", cr.Namespace, key.AWSClusterConfigName(cr)))

	presentAWSClusterConfig, err := r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cr.Namespace).Get(key.AWSClusterConfigName(cr), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find AWSClusterConfig %#q/%#q", cr.Namespace, key.AWSClusterConfigName(cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating AWSClusterConfig %#q/%#q", cr.Namespace, key.AWSClusterConfigName(cr)))

		newAWSClusterConfig := r.constructAWSClusterConfig(cr, machineDeployments, versionBundles)

		_, err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cr.Namespace).Create(&newAWSClusterConfig)
		if errors.IsAlreadyExists(err) {
			r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("AWSClusterConfig %#q/%#q already exists", newAWSClusterConfig.Namespace, newAWSClusterConfig.Name))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created AWSClusterConfig %#q/%#q", newAWSClusterConfig.Namespace, newAWSClusterConfig.Name))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found AWSClusterConfig %#q/%#q", cr.Namespace, presentAWSClusterConfig.Name))

	// Map desired state from Cluster to AWSClusterConfig.
	newAWSClusterConfig := r.mapClusterToAWSClusterConfig(*presentAWSClusterConfig, cr, machineDeployments, versionBundles)

	if reflect.DeepEqual(presentAWSClusterConfig.Spec, newAWSClusterConfig.Spec) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("current AWSClusterConfig %#q/%#q is up-to-date; no update needed.", newAWSClusterConfig.Namespace, newAWSClusterConfig.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating AWSClusterConfig %#q/%#q", newAWSClusterConfig.Namespace, newAWSClusterConfig.Name))

	_, err = r.g8sClient.CoreV1alpha1().AWSClusterConfigs(cr.Namespace).Update(&newAWSClusterConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated AWSClusterConfig %#q/%#q", newAWSClusterConfig.Namespace, newAWSClusterConfig.Name))

	return nil
}

func (r *Resource) mapClusterToAWSClusterConfig(awsClusterConfig corev1alpha1.AWSClusterConfig, cr clusterv1alpha1.Cluster, machineDeployments []clusterv1alpha1.MachineDeployment, versionBundles []versionbundle.Bundle) corev1alpha1.AWSClusterConfig {
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.AvailabilityZones = NumberOfAZsWithNodePools
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.DNSZone = key.ClusterDNSZone(cr)
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.ID = key.ClusterID(&cr)
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.Name = key.ClusterName(cr)
	awsClusterConfig.Spec.Guest.ClusterGuestConfig.ReleaseVersion = key.ReleaseVersion(&cr)

	var awsClusterConfigVersionBundle corev1alpha1.AWSClusterConfigSpecVersionBundle
	var transformedVBs []corev1alpha1.ClusterGuestConfigVersionBundle
	{
		for _, b := range versionBundles {
			bundle := corev1alpha1.ClusterGuestConfigVersionBundle{
				Name:    b.Name,
				Version: b.Version,
			}

			transformedVBs = append(transformedVBs, bundle)

			if b.Name == "cluster-operator" {
				awsClusterConfigVersionBundle.Version = b.Version
			}
		}
	}

	awsClusterConfig.Spec.Guest.ClusterGuestConfig.VersionBundles = transformedVBs
	awsClusterConfig.Spec.Guest.CredentialSecret.Name = key.ClusterCredentialSecretName(cr)
	awsClusterConfig.Spec.Guest.CredentialSecret.Namespace = key.ClusterCredentialSecretNamespace(cr)
	awsClusterConfig.Spec.VersionBundle = awsClusterConfigVersionBundle

	awsClusterConfig.Spec.Guest.Masters = []corev1alpha1.AWSClusterConfigSpecGuestMaster{
		{
			AWSClusterConfigSpecGuestNode: corev1alpha1.AWSClusterConfigSpecGuestNode{
				InstanceType: key.ClusterMasterInstanceType(cr),
			},
		},
	}

	var workers []corev1alpha1.AWSClusterConfigSpecGuestWorker
	for _, md := range machineDeployments {
		for i := 0; i < int(md.Status.Replicas); i++ {
			w := corev1alpha1.AWSClusterConfigSpecGuestWorker{
				AWSClusterConfigSpecGuestNode: corev1alpha1.AWSClusterConfigSpecGuestNode{
					InstanceType: key.MachineDeploymentWorkerInstanceType(md),
				},
				Labels: map[string]string{
					label.Cluster:           key.ClusterID(&cr),
					label.ReleaseVersion:    key.ReleaseVersion(&md),
					label.MachineDeployment: key.MachineDeployment(&md),
				},
			}
			workers = append(workers, w)
		}
	}
	awsClusterConfig.Spec.Guest.Workers = workers

	return awsClusterConfig
}

func (r *Resource) constructAWSClusterConfig(cr clusterv1alpha1.Cluster, machineDeployments []clusterv1alpha1.MachineDeployment, versionBundles []versionbundle.Bundle) corev1alpha1.AWSClusterConfig {
	cc := corev1alpha1.AWSClusterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSClusterConfig",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.AWSClusterConfigName(cr),
			Namespace: cr.Namespace,
		},
	}

	return r.mapClusterToAWSClusterConfig(cc, cr, machineDeployments, versionBundles)
}
