package response

type Workers struct {
	CountPerCluster CountPerCluster `json:"count_per_cluster"`
	InstanceType    InstanceType    `json:"instance_type,omitempty"`
}
