package tenantclients

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v17/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var cluster v1alpha1.Cluster
	var err error

	// This resource is used both from Cluster and MachineDeployment
	// controllers so it must work with both types.
	switch obj.(type) {
	case *v1alpha1.Cluster:
		cluster, err = key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

	case *v1alpha1.MachineDeployment:
		cr, err := key.ToMachineDeployment(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		m, err := r.cmaClient.ClusterV1alpha1().Clusters(corev1.NamespaceAll).Get(key.ClusterID(&cr), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		cluster = *m

	default:
		return microerror.Maskf(wrongTypeError, "expected '%T' or '%T', got '%T'", &v1alpha1.Cluster{}, &v1alpha1.MachineDeployment{}, obj)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var g8sClient versioned.Interface
	var k8sClient kubernetes.Interface
	{
		g8sClient, err = r.tenant.NewG8sClient(ctx, key.ClusterID(&cluster), key.ClusterAPIEndpoint(cluster))
		if err != nil {
			return microerror.Mask(err)
		}

		k8sClient, err = r.tenant.NewK8sClient(ctx, key.ClusterID(&cluster), key.ClusterAPIEndpoint(cluster))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		cc.Client.TenantCluster.G8s = g8sClient
		cc.Client.TenantCluster.K8s = k8sClient
	}

	return nil
}
