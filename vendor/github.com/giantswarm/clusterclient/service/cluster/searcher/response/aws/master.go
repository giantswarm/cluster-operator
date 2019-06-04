package aws

// Master configures AWS-specific worker node settings.
type Master struct {
	InstanceType string `json:"instance_type"`
}

// DefaultMaster provides default Master.
func DefaultMaster() Master {
	return Master{
		InstanceType: "",
	}
}
