package awsconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/service/controller/aws/v5/key"
)

// GetCurrentState takes observed custom object as an input and based on that
// information looks for current state of AWSConfig and returns it. Return
// value is of type *v1alpha1.AWSConfig.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	awsClusterConfig, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	awsConfigName := key.AWSConfigName(awsClusterConfig)

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for awsconfig in the Kubernetes API", "awsConfigName", awsConfigName)

	awsConfig, err := r.g8sClient.ProviderV1alpha1().AWSConfigs(awsClusterConfig.Namespace).Get(awsConfigName, apismetav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find an awsconfig in the Kubernetes API", "awsConfigName", awsConfigName)
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found an awsconfig in the Kubernetes API", "awsConfigName", awsConfigName)

	return awsConfig, nil
}
