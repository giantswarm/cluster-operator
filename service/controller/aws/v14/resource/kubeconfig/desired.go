package kubeconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/kubeconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v14/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v14/key"
)

func (r *StateGetter) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	customObject, err := awskey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := awskey.ClusterGuestConfig(customObject)
	apiDomain, err := key.APIDomain(clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	appOperator, err := r.certsSearcher.SearchAppOperator(clusterGuestConfig.ID)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	c := k8srestconfig.Config{
		Logger: r.logger,

		Address:   apiDomain,
		InCluster: false,
		TLS: k8srestconfig.ConfigTLS{
			CAData:  appOperator.APIServer.CA,
			CrtData: appOperator.APIServer.Crt,
			KeyData: appOperator.APIServer.Key,
		},
	}
	restConfig, err := k8srestconfig.New(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	yamlBytes, err := kubeconfig.NewKubeConfigForRESTConfig(ctx, restConfig, fmt.Sprintf("giantswarm-%s", clusterGuestConfig.ID), "")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretName := fmt.Sprintf("%s-kubeconfig", clusterGuestConfig.ID)

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Labels: map[string]string{
				label.ManagedBy: r.projectName,
			},
		},
		Data: map[string][]byte{
			"kubeConfig": yamlBytes,
		},
	}

	return []*corev1.Secret{&secret}, nil
}
