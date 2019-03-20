package kubeconfig

import (
	"context"
	"fmt"
	"github.com/giantswarm/cluster-operator/pkg/v13/chartconfig"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *StateGetter) GetCurrentState(ctx context.Context, clusterConfig chartconfig.ClusterConfig) ([]*corev1.Secret, error) {
	secretName := fmt.Sprintf("%s-kubeconfig", clusterConfig.ClusterID)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding kubeconfig secret %#q", secretName))

	secret, err := r.k8sClient.CoreV1().Secrets(r.resourceNamespace).Get(secretName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find kubeconfig secret %#q", secretName))
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found kubeconfig resource %#q", secretName))

	return []*corev1.Secret{secret}, nil
}
