// +build k8srequired

package env

import (
	"os"
)

const (
	// EnvVarClusterID is the process environment variable representing the
	// CLUSTER_NAME env var.
	//
	// TODO rename to CLUSTER_ID. Note this also had to be changed in the
	// framework package of e2e-harness.
	EnvVarClusterID = "CLUSTER_NAME"
)

var (
	clusterID string
)

func init() {
	// NOTE that implications of changing the order of initialization here means
	// breaking the initialization behaviour.
	clusterID = os.Getenv(EnvVarClusterID)
	if clusterID == "" {
		os.Setenv(EnvVarClusterID, ClusterID())
	}
}

func ClusterID() string {
	return clusterID
}
