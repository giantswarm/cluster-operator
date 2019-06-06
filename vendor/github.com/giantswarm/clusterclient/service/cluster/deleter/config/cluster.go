package config

type Cluster struct {
	ID string `json:"id"`
}

func DefaultCluster() *Cluster {
	return &Cluster{
		ID: "",
	}
}
