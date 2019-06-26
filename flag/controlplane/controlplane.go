package controlplane

// ControlPlane represents control plane specific configuration,
// used for templating resources in tenant clusters.
type ControlPlane struct {
	// Subnets is a list of control-plane node subnets,
	// where Prometheus might be running.
	// For clouds it is worker subnets, for on-prem - whole private subnet.
	Subnets string
}
