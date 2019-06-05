package kvm

// ProtocolPort represents a mapping from a port on the host cluster to the ingress
// of a guest cluster for a given protocol.
type ProtocolPort struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}
