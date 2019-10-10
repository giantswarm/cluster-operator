package encryptionkey

import (
	"fmt"

	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v21/key"
)

func secretName(cr v1alpha1.Cluster) string {
	return fmt.Sprintf("%s-%s", key.ClusterID(&cr), "encryption")
}
