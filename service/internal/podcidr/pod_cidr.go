package podcidr

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bootstrapkubeadmv1alpha3 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/internal/podcidr/internal/cache"
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
		client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
	)
	if err != nil {
		return bootstrapkubeadmv1alpha3.KubeadmConfig, microerror.Mask(err)
	}

	if len(list.Items) == 0 {
		return bootstrapkubeadmv1alpha3.KubeadmConfig, microerror.Mask(notFoundError)
	}
	if len(list.Items) > 1 {
		return bootstrapkubeadmv1alpha3.KubeadmConfig, microerror.Mask(tooManyCRsError)
	}

	return list.Items[0], nil
}
