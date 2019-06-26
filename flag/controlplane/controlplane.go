package controlplane

// ControlPlane represents control plane specific configuration,
// used for templating resources in tenant clusters.
type ControlPlane struct {
	// WorkerSubnets is comma-separated list of control-plane
	// workers subnets.
	WorkerSubnets string
}
