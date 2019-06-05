package request

// Memory configures the machine memory.
type Memory struct {
	SizeGB float64 `json:"size_gb"`
}

// DefaultMemory provides a default ram configuration by best effort.
func DefaultMemory() Memory {
	return Memory{
		SizeGB: 0,
	}
}
