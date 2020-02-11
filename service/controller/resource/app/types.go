package app

type defaultConfig struct {
	Catalog         string `json:"catalog"`
	Namespace       string `json:"namespace"`
	UseUpgradeForce bool   `json:"useUpgradeForce"`
}

type overrideConfig map[string]overrideProperties

type overrideProperties struct {
	Chart           string `json:"chart"`
	Namespace       string `json:"namespace"`
	UseUpgradeForce *bool  `json:"useUpgradeForce,omitempty"`
}
