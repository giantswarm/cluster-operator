package response

type AvailabilityZones struct {
	Default int      `json:"default"`
	Max     int      `json:"max"`
	Zones   []string `json:"zones,omitempty"`
}
