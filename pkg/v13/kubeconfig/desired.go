package kubeconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/kubeconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *StateGetter) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	cr, err := ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	appOperator, err := r.certsSearcher.SearchAppOperator(cr.ClusterID)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	c := k8srestconfig.Config{
		Logger: r.logger,

		Address:   cr.APIDomain,
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

	yamlBytes, err := kubeconfig.NewKubeConfigForRESTConfig(ctx, restConfig, fmt.Sprintf("giantswarm-%s", cr.ClusterID), "")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.resourceName,
			Namespace: r.resourceNamespace,
		},
		Data: map[string][]byte{
			"kubeConfig": yamlBytes,
		},
	}

	return []*corev1.Secret{&secret}, nil
}
