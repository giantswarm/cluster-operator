package provider

// Provider data structure holds infrastructure provider specific configuration
// flags.
type Provider struct {
	// Kind contains infrastructure provider type. It can be 'aws', 'azure' or
	// 'kvm'.
	Kind string
}
