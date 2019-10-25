package response

type General struct {
	AvailabilityZones AvailabilityZones `json:"availability_zones"`
	Datacenter        string            `json:"datacenter"`
	InstallationName  string            `json:"installation_name"`
	Provider          string            `json:"provider"`
}
