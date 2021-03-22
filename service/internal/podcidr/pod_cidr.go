package podcidr

import (
	"context"
	"strings"

	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bootstrapkubeadmv1alpha3 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	K8sClient k8sclient.Interface

	InstallationCIDR string
}

type PodCIDR struct {
	k8sClient k8sclient.Interface

	installationCIDR string
}

func New(c Config) (*PodCIDR, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	if c.InstallationCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationCIDR must not be empty", c)
	}

	p := &PodCIDR{
		k8sClient: c.K8sClient,

		installationCIDR: c.InstallationCIDR,
	}

	return p, nil
}

func (p *PodCIDR) PodCIDR(ctx context.Context, obj interface{}) (string, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	cl, err := p.lookupCluster(ctx, cr)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var podCIDR string
	if cl.Spec.ClusterConfiguration.Networking.PodSubnet == "" {
		podCIDR = p.installationCIDR
	} else {
		podCIDR = cl.Spec.ClusterConfiguration.Networking.PodSubnet
	}

	return podCIDR, nil
}

func (p *PodCIDR) lookupCluster(ctx context.Context, cr metav1.Object) (bootstrapkubeadmv1alpha3.KubeadmConfig, error) {
	var list bootstrapkubeadmv1alpha3.KubeadmConfigList

	err := p.k8sClient.CtrlClient().List(
		ctx,
		&list,
		client.InNamespace(cr.GetNamespace()),
	)
	if err != nil {
		return bootstrapkubeadmv1alpha3.KubeadmConfig{}, microerror.Mask(err)
	}

	filteredItems := filterByName(list.Items, "control-plane")

	if len(filteredItems) == 0 {
		return bootstrapkubeadmv1alpha3.KubeadmConfig{}, microerror.Mask(notFoundError)
	}

	if len(filteredItems) > 1 {
		return bootstrapkubeadmv1alpha3.KubeadmConfig{}, microerror.Mask(tooManyCRsError)
	}

	return filteredItems[0], nil
}

func filterByName(items []bootstrapkubeadmv1alpha3.KubeadmConfig, substr string) (result []bootstrapkubeadmv1alpha3.KubeadmConfig) {
	for _, item := range items {
		if strings.Contains(item.Name, substr) {
			result = append(result, item)
		}
	}

	return result
}
