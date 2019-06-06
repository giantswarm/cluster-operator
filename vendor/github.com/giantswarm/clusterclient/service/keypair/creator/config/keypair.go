package config

type KeyPair struct {
	CommonNamePrefix string `json:"cn_prefix"`
	Description      string `json:"description"`
	Organizations    string `json:"certificate_organizations"`
	TTL              int    `json:"ttl"`
}

func DefaultKeyPair() *KeyPair {
	return &KeyPair{
		CommonNamePrefix: "",
		Description:      "",
		Organizations:    "",
		TTL:              0,
	}
}
