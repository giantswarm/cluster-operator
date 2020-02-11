package kvm

// Cluster configures KVM-specific cluster settings.
type Cluster struct {
	PortMappings []ProtocolPort `json:"port_mappings,omitempty"`
}
