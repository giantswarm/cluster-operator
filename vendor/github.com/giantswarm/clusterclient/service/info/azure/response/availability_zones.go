package response

type AvailabilityZones struct {
	DefaultCount int   `json:"default"`
	MaxCount     int   `json:"max"`
	Zones        []int `json:"zones,omitempty"`
}
