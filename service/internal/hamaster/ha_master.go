package hamaster

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

type Config struct {
	K8sClient k8sclient.Interface

	Provider string
}

type HAMaster struct {
	k8sClient k8sclient.Interface

	provider string
}

func New(config Config) (*HAMaster, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	h := &HAMaster{
		k8sClient: config.K8sClient,

		provider: config.Provider,
	}

	return h, nil
}

func (h *HAMaster) Enabled(ctx context.Context, cluster string) (bool, error) {
	if h.provider != label.ProviderAWS {
		return false, nil
	}

	var list infrastructurev1alpha2.G8sControlPlaneList

	err := h.k8sClient.CtrlClient().List(
		ctx,
		&list,
		client.MatchingLabels{label.Cluster: cluster},
	)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if len(list.Items) == 0 {
		return false, microerror.Mask(notFoundError)
	}

	if key.G8sControlPlaneReplicas(list.Items[0]) == 1 {
		return false, nil
	}

	return true, nil
}
