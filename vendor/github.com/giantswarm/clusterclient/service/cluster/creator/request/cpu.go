package request

// CPU configures the machine CPU.
type CPU struct {
	Cores int `json:"cores"`
}

// DefaultCPU provides a default CPU configuration by best effort.
func DefaultCPU() CPU {
	return CPU{
		Cores: 0,
	}
}
