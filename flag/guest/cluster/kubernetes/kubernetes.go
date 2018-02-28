package kubernetes

// Kubernetes is a data structure to hold guest cluster Kubernetes specific
// configuration flags.
type Kubernetes struct {
	API               API
	Domain            string
	Hyperkube         Hyperkube
	IngressController IngressController
	Kubelet           Kubelet
	NetworkSetup      NetworkSetup
	SSH               SSH
}

// API is a data structure to hold guest cluster Kubernetes API specific
// configuration flags.
type API struct {
	AltNames       string
	ClusterIPRange string
	InsecurePort   int
	SecurePort     int
}

// Hyperkube is a data structure to hold guest cluster Kubernetes Hyperkube
// image specific configuration flags.
type Hyperkube struct {
	Docker Docker
}

// IngressController is a data structure to hold guest cluster ingress
// controller specific configuration flags.
type IngressController struct {
	Docker Docker
}

// Kubelet is a data structure to hold guest cluster kubelet specific
// configuration flags.
type Kubelet struct {
	AltNames string
	Labels   string
	Port     int
}

// NetworkSetup is a data structure to hold guest cluster network setup
// configuration flags.
type NetworkSetup struct {
	Docker Docker
}

// Docker is a data structure to hold Docker image configuration flag.
type Docker struct {
	Image string
}

// SSH is a data structure to hold guest cluster SSH specific configuration
// flags.
type SSH struct {
	UserList []SSHUser
}

// SSHUser is a data structure to represent username and public key pair.
type SSHUser struct {
	Name      string
	PublicKey string
}
