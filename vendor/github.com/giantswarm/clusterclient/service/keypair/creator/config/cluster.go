package config

type Cluster struct {
	ID string
}

func DefaultCluster() *Cluster {
	return &Cluster{
		ID: "",
	}
}
