package response

type CountPerCluster struct {
	Max     int `json:"max"`
	Default int `json:"default"`
}

func DefaultCountPerCluster() CountPerCluster {
	return CountPerCluster{
		Max:     0,
		Default: 0,
	}
}
