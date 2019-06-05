package response

type General struct {
	AvailabilityZones AvailabilityZones `json:"availability_zones"`
	InstallationName  string            `json:"installation_name"`
	Provider          string            `json:"provider"`

	// To be implemented:
	// datacenter
}
