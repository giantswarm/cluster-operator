package kubeconfig

import (
	"context"

	"github.com/giantswarm/kubeconfig/v2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster/v3/pkg/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	bd, err := r.baseDomain.BaseDomain(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	var restConfig *rest.Config
	{
		restConfig, err = r.tenant.NewRestConfig(ctx, key.ClusterID(&cr), key.KubeConfigEndpoint(&cr, bd))
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var secret *corev1.Secret
	{
		b, err := kubeconfig.NewKubeConfigForRESTConfig(ctx, restConfig, key.KubeConfigClusterName(&cr), "")
		if err != nil {
			return nil, microerror.Mask(err)
		}

		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.KubeConfigSecretName(&cr),
				Namespace: key.ClusterID(&cr),
				Labels: map[string]string{
					label.Cluster:      key.ClusterID(&cr),
					label.ManagedBy:    project.Name(),
					label.Organization: key.OrganizationID(&cr),
					label.ServiceType:  label.ServiceTypeManaged,
				},
			},
			Data: map[string][]byte{
				"kubeConfig": b,
			},
		}
	}

	return []*corev1.Secret{secret}, nil
}
