package response

type General struct {
	InstallationName string `json:"installation_name"`
	Provider         string `json:"provider"`

	// To be implemented:
	// datacenter
}

func DefaultGeneral() General {
	return General{
		InstallationName: "",
		Provider:         "",
	}
}
