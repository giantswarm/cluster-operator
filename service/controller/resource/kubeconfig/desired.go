package kubeconfig

import (
	"context"

	"github.com/giantswarm/kubeconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster/v2/pkg/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if cc.Status.Endpoint.Base == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no endpoint base in controller context yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil, nil
	}

	var restConfig *rest.Config
	{
		restConfig, err = r.tenant.NewRestConfig(ctx, key.ClusterID(&cr), key.KubeConfigEndpoint(cr, cc.Status.Endpoint.Base))
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
