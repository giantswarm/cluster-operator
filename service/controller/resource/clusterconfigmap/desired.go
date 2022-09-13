package clusterconfigmap

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	k8smetadata "github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/cluster-operator/v4/pkg/annotation"
	"github.com/giantswarm/cluster-operator/v4/pkg/label"
	"github.com/giantswarm/cluster-operator/v4/pkg/project"
	"github.com/giantswarm/cluster-operator/v4/service/controller/key"
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
	{
		if r.provider == "aws" {
			useProxyProtocol = true
		}
	}

	var irsa bool
	var accountID string
	{
		if r.provider == "aws" {

			awsCluster := &v1alpha3.AWSCluster{}
			err := r.ctrlClient.Get(ctx, types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, awsCluster)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			if _, ok := awsCluster.Annotations[k8smetadata.AWSIRSA]; ok {
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
		}
	}

	configMapSpecs := []configMapSpec{
		{
			Name:      key.ClusterConfigMapName(&cr),
			Namespace: key.ClusterID(&cr),
			Values: map[string]interface{}{
				"aws": map[string]interface{}{
					"accountID": accountID,
					"irsa":      strconv.FormatBool(irsa),
				},
				"baseDomain": key.TenantEndpoint(&cr, bd),
				"chartOperator": map[string]interface{}{
					"cni": map[string]interface{}{
						"install": true,
					},
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
				"organization": key.OrganizationID(&cr),
				"remoteWrite": []map[string]interface{}{{
					"url": remoteWriteUrl(bd, key.ClusterID(&cr)),
				}},
			},
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
			Values: map[string]interface{}{
				"defaultPolicies": map[string]interface{}{
					"enabled": true,
				},
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

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapSpec.Name,
			Namespace: configMapSpec.Namespace,
			Annotations: map[string]string{
				annotation.Notes: fmt.Sprintf("DO NOT EDIT. Values managed by %s.", project.Name()),
			},
			Labels: map[string]string{
				label.Cluster:      key.ClusterID(&cr),
				label.ManagedBy:    project.Name(),
				label.Organization: key.OrganizationID(&cr),
				label.ServiceType:  label.ServiceTypeManaged,
			},
		},
		Data: map[string]string{
			"values": string(yamlValues),
		},
	}

	return cm, nil
}

func remoteWriteUrl(baseDomain, clusterID string) string {
	return fmt.Sprintf("https://prometheus.g8s.%s/%s/api/v1/write", baseDomain, clusterID)
}
