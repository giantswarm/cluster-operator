package kubelet

// Kubelet is a data structure to hold guest cluster kubelet specific
// configuration flags.
type Kubelet struct {
	AltNames string
	Labels   string
	Port     string
}
