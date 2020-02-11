package response

type AvailabilityZones struct {
	DefaultCount int      `json:"default"`
	MaxCount     int      `json:"max"`
	Zones        []string `json:"zones,omitempty"`
}
