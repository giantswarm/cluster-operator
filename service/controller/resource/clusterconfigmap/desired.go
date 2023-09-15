package clusterconfigmap

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	releasev1alpha1 "github.com/giantswarm/release-operator/v4/api/v1alpha1"

	"github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/cluster-operator/v5/pkg/annotation"
	"github.com/giantswarm/cluster-operator/v5/pkg/label"
	"github.com/giantswarm/cluster-operator/v5/pkg/project"
	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	bd, err := r.baseDomain.BaseDomain(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusterCA string
	{
		apiSecret, err := r.k8sClient.CoreV1().Secrets(cr.Namespace).Get(ctx, key.APISecretName(&cr), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// During cluster creation there may be a delay until the
			// cert is issued.
			r.logger.Debugf(ctx, "secret '%s/%s' not found cannot set cluster CA", cr.Namespace, key.APISecretName(&cr))
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterCA = string(apiSecret.Data["ca"])
	}

	var podCIDR string
	{
		podCIDR, err = r.podCIDR.PodCIDR(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// useProxyProtocol is only enabled by default for AWS clusters.
	var useProxyProtocol bool
	// enableCiliumNetworkPolicy is only enabled by default for AWS clusters.
	var enableCiliumNetworkPolicy bool
	{
		if key.IsAWS(r.provider) {
			useProxyProtocol = true
			enableCiliumNetworkPolicy = true
		}
	}

	values := map[string]interface{}{
		"baseDomain": key.TenantEndpoint(&cr, bd),
		"bootstrapMode": map[string]interface{}{
			"enabled": true,
		},
		"cluster": map[string]interface{}{
			"calico": map[string]interface{}{
				"CIDR": podCIDR,
			},
			"kubernetes": map[string]interface{}{
				"API": map[string]interface{}{
					"clusterIPRange": r.clusterIPRange,
				},
				"DNS": map[string]interface{}{
					"IP": r.dnsIP,
				},
			},
		},
		"clusterCA":    clusterCA,
		"clusterDNSIP": r.dnsIP,
		"clusterID":    key.ClusterID(&cr),
		"ciliumNetworkPolicy": map[string]interface{}{
			"enabled": enableCiliumNetworkPolicy,
		},
		"global": map[string]interface{}{
			"podSecurityStandards": map[string]interface{}{
				"enforced": true,
			},
		},
	}

	externalDnsValues := map[string]interface{}{
		"txtOwnerId":       "giantswarm-io-external-dns",
		"txtPrefix":        key.ClusterID(&cr),
		"annotationFilter": "giantswarm.io/external-dns=managed",
		"sources": []string{
			"service",
		},
	}

	if key.IsAWS(r.provider) {
		var irsa bool
		var accountID string
		var vpcID string

		awsCluster := &v1alpha3.AWSCluster{}
		err := r.ctrlClient.Get(ctx, types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, awsCluster)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if key.IRSAEnabled(awsCluster) {
			irsa = true
		}

		secret, err := r.k8sClient.CoreV1().Secrets(awsCluster.Spec.Provider.CredentialSecret.Namespace).Get(ctx, awsCluster.Spec.Provider.CredentialSecret.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "secret '%s/%s' not found cannot set accountID", cr.Namespace, key.APISecretName(&cr))
			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
		arn := string(secret.Data["aws.awsoperator.arn"])
		if arn == "" {
			return nil, microerror.Mask(fmt.Errorf("Unable to find ARN from secret %s/%s", secret.Namespace, secret.Name))
		}

		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		accountID = re.FindAllString(arn, 1)[0]

		vpcID = awsCluster.Status.Provider.Network.VPCID

		values["aws"] = map[string]interface{}{
			"accountID": accountID,
			"irsa":      strconv.FormatBool(irsa),
			"region":    awsCluster.Spec.Provider.Region,
			"vpcID":     vpcID,
		}

		externalDnsValues["extraArgs"] = []string{
			"--aws-batch-change-interval=10s",
		}
		externalDnsValues["aws"] = map[string]interface{}{
			"batchChangeInterval": "null",
		}
		externalDnsValues["serviceAccount"] = map[string]interface{}{
			"annotations": map[string]interface{}{
				"eks.amazonaws.com/role-arn": fmt.Sprintf("arn:aws:iam::%s:role/%s-Route53Manager-Role", accountID, key.ClusterID(&cr)),
			},
		}
		externalDnsValues["domainFilters"] = []string{
			key.TenantEndpoint(&cr, bd),
		}
	}

	ciliumValues := map[string]interface{}{
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"cni": map[string]interface{}{
			"exclusive": false,
		},
		"extraEnv": []map[string]string{
			{
				"name":  "CNI_CONF_NAME",
				"value": "21-cilium.conf",
			},
		},
	}

	// We only need this if the cluster is in overlay mode during the upgrade
	if key.ForceDisableCiliumKubeProxyReplacement(cr) && !key.CiliumEniModeEnabled(cr) {
		ciliumValues["kubeProxyReplacement"] = "disabled"
	} else {
		ciliumValues["kubeProxyReplacement"] = "strict"
		ciliumValues["k8sServiceHost"] = key.APIEndpoint(&cr, bd)
		ciliumValues["k8sServicePort"] = "443"
		ciliumValues["cleanupKubeProxy"] = true
	}

	if key.IsAWS(r.provider) && key.CiliumEniModeEnabled(cr) {
		// Add selector to not interfere with nodes still running in AWS CNI
		awsCluster := &v1alpha3.AWSCluster{}
		err := r.ctrlClient.Get(ctx, types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, awsCluster)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var re releasev1alpha1.Release
		err = r.ctrlClient.Get(
			ctx,
			types.NamespacedName{Name: key.ReleaseName(key.ReleaseVersion(&cr))},
			&re,
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var awsOperatorRelease string
		for _, v := range re.Spec.Components {
			if v.Name == "aws-operator" {
				awsOperatorRelease = v.Version
			}
		}

		if awsOperatorRelease == "" {
			return nil, microerror.Mask(releaseNotFound)
		}

		// This is a hack to only introduce the selector during the upgrade on the new nodes, old ones work with AWS CNI
		if key.ForceDisableCiliumKubeProxyReplacement(cr) {
			ciliumValues["nodeSelector"] = map[string]interface{}{
				"aws-operator.giantswarm.io/version": awsOperatorRelease,
			}
		}

		ciliumValues["eni"] = map[string]interface{}{
			"enabled": true,
			//"awsEnablePrefixDelegation": true,
		}

		ciliumValues["ipam"] = map[string]interface{}{
			"mode": "eni",
		}

		// there is autodiscoverability on the VPC CIDrs
		ciliumValues["ipv4NativeRoutingCIDR"] = podCIDR

		// https://docs.cilium.io/en/v1.13/network/concepts/routing/#id5
		ciliumValues["endpointRoutes"] = map[string]interface{}{
			"enabled": true,
		}

		ciliumValues["operator"] = map[string]interface{}{
			"extraArgs": []string{
				"--aws-release-excess-ips=true",
			},
		}

		ciliumValues["egressMasqueradeInterfaces"] = "eth+"
		ciliumValues["tunnel"] = "disabled"
		// Used by cilium to tag ENIs it creates and be able to filter and clean them up.
		ciliumValues["cluster"] = map[string]interface{}{
			"name": key.ClusterID(&cr),
		}
		ciliumValues["cni"] = map[string]interface{}{
			"customConf": true,
			"exclusive":  true,
			"configMap":  "cilium-cni-configuration",
		}
		ciliumValues["extraEnv"] = []map[string]string{
			{
				"name":  "CNI_CONF_NAME",
				"value": "21-cilium.conflist",
			},
		}

	}

	configMapSpecs := []configMapSpec{
		{
			Name:      key.ClusterConfigMapName(&cr),
			Namespace: key.ClusterID(&cr),
			Values:    values,
		},
		{
			Name:      "ingress-controller-values",
			Namespace: key.ClusterID(&cr),
			Values: map[string]interface{}{
				"baseDomain": key.TenantEndpoint(&cr, bd),
				"clusterID":  key.ClusterID(&cr),
				"configmap": map[string]interface{}{
					"use-proxy-protocol": strconv.FormatBool(useProxyProtocol),
				},
			},
		},
		{
			Name:      "cilium-user-values",
			Namespace: key.ClusterID(&cr),
			Values:    ciliumValues,
		},
		{
			Name:      "external-dns-cluster-values",
			Namespace: key.ClusterID(&cr),
			Values:    externalDnsValues,
			Labels: map[string]string{
				"app.kubernetes.io/name": "external-dns",
			},
			Annotations: map[string]string{
				"cluster-operator.giantswarm.io/app-config-priority": "130",
			},
		},
	}

	var configMaps []*corev1.ConfigMap

	for _, spec := range configMapSpecs {
		configMap, err := newConfigMap(cr, spec)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

func newConfigMap(cr apiv1beta1.Cluster, configMapSpec configMapSpec) (*corev1.ConfigMap, error) {
	yamlValues, err := yaml.Marshal(configMapSpec.Values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	annotations := map[string]string{
		annotation.Notes: fmt.Sprintf("DO NOT EDIT. Values managed by %s.", project.Name()),
	}
	for k, v := range configMapSpec.Annotations {
		annotations[k] = v
	}

	labels := map[string]string{
		label.Cluster:      key.ClusterID(&cr),
		label.ManagedBy:    project.Name(),
		label.Organization: key.OrganizationID(&cr),
		label.ServiceType:  label.ServiceTypeManaged,
	}
	for k, v := range configMapSpec.Labels {
		labels[k] = v
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        configMapSpec.Name,
			Namespace:   configMapSpec.Namespace,
			Annotations: annotations,
			Labels:      labels,
		},
		Data: map[string]string{
			"values": string(yamlValues),
		},
	}

	return cm, nil
}
