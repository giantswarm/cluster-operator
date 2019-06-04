package config

type Organization struct {
	ID string
}

func DefaultOrganization() *Organization {
	return &Organization{
		ID: "",
	}
}
