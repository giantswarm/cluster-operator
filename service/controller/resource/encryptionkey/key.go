package encryptionkey

import (
	"fmt"

	clusterv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/cluster-operator/service/controller/key"
)

func secretName(cr clusterv1alpha2.Cluster) string {
	return fmt.Sprintf("%s-%s", key.ClusterID(&cr), "encryption")
}
