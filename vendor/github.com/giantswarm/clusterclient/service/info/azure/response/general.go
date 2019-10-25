package response

type General struct {
	Datacenter       string `json:"datacenter"`
	InstallationName string `json:"installation_name"`
	Provider         string `json:"provider"`
}

func DefaultGeneral() General {
	return General{
		Datacenter:       "",
		InstallationName: "",
		Provider:         "",
	}
}
