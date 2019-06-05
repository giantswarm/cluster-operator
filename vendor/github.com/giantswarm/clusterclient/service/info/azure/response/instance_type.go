package response

type VMSize struct {
	Options []string `json:"options"`
	Default string   `json:"default"`
}

func DefaultVMSize() VMSize {
	return VMSize{
		Options: []string{},
		Default: "",
	}
}
