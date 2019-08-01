package app

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v18/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v18/key"
)

func (s StateGetter) GetCurrentState(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error) {
	customObject, err := awskey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterGuestConfig := awskey.ClusterGuestConfig(customObject)
	clusterID := key.ClusterID(clusterGuestConfig)

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, s.projectName),
	}

	appList, err := s.g8sClient.ApplicationV1alpha1().Apps(clusterID).List(listOptions)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	apps := make([]*v1alpha1.App, 0, len(appList.Items))

	for _, item := range appList.Items {
		item := item
		apps = append(apps, &item)
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d apps in the cluster", len(apps)))

	return apps, nil
}
