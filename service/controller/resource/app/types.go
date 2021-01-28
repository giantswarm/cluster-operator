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

type Index struct {
	Entries map[string][]IndexEntry `json:"entries"`
}

type IndexEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
