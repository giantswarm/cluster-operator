package label

const (
	// MasterNode label denotes a K8s cluster master node.
	MasterNode = "node.kubernetes.io/master"
	// MasterNodeRole label denotes K8s cluster master node role.
	MasterNodeRole = "node-role.kubernetes.io/master"
)
