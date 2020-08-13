package encryptionkey

import (
	"fmt"

	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func secretName(cr apiv1alpha2.Cluster) string {
	return fmt.Sprintf("%s-%s", key.ClusterID(&cr), "encryption")
}
