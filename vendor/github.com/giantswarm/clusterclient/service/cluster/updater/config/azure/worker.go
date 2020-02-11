package azure

// Worker configures Azure-specific worker node settings.
type Worker struct {
	VmSize string `json:"vm_size"`
}
