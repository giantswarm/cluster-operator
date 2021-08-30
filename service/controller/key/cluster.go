package key

import (
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func APIEndpoint(getter LabelsGetter, base string) string {
	return fmt.Sprintf("api.%s.k8s.%s", ClusterID(getter), base)
}

func KubeConfigEndpoint(getter LabelsGetter, base string) string {
	return fmt.Sprintf("https://%s", APIEndpoint(getter, base))
}

func TenantEndpoint(getter LabelsGetter, base string) string {
	return fmt.Sprintf("%s.k8s.%s", ClusterID(getter), base)
}

func ToCluster(v interface{}) (apiv1alpha3.Cluster, error) {
	if v == nil {
		return apiv1alpha3.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1alpha3.Cluster{}, v)
	}

	p, ok := v.(*apiv1alpha3.Cluster)
	if !ok {
		return apiv1alpha3.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1alpha3.Cluster{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
