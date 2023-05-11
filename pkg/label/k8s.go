package label

const (
	MasterNodeRole = "node-role.kubernetes.io/master"
	// WorkerNodeRole label denotes K8s cluster worker node role.
	WorkerNodeRole = "node-role.kubernetes.io/worker"
)

// MasterNodeRoles labels denote K8s cluster master node role.
var MasterNodeRoles = []string{
	"node-role.kubernetes.io/master",
	"node-role.kubernetes.io/control-plane",
}
