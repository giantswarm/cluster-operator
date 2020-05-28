package label

const (
	// MasterNodeRole label denotes K8s cluster master node role.
	MasterNodeRole = "node-role.kubernetes.io/master"
	// WorkerNodeRole label denotes K8s cluster worker node role.
	WorkerNodeRole = "node-role.kubernetes.io/worker"
)
