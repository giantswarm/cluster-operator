package kubeconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/kubeconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v23/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var appOperator certs.AppOperator
	{
		appOperator, err = r.certsSearcher.SearchAppOperator(key.ClusterID(&cr))
		if certs.IsTimeout(err) {
			// We can't continue without the app-operator api certs. We will retry
			// during the next execution.
			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: r.logger,

			Address:   fmt.Sprintf("https://%s", key.ClusterAPIEndpoint(cr)),
			InCluster: false,
			TLS: k8srestconfig.ConfigTLS{
				CAData:  appOperator.APIServer.CA,
				CrtData: appOperator.APIServer.Crt,
				KeyData: appOperator.APIServer.Key,
			},
		}

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
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
