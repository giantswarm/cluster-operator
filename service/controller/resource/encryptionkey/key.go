package encryptionkey

import (
	"fmt"

	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func secretName(cr apiv1alpha3.Cluster) string {
	return fmt.Sprintf("%s-%s", key.ClusterID(&cr), "encryption")
}
