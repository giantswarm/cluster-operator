package v1alpha1

type ClusterGuestConfig struct {
	CommonDomain   string                            `json:"commonDomain" yaml:"commonDomain"`
	ID             string                            `json:"id" yaml:"id"`
	Name           string                            `json:"name,omitempty" yaml:"name,omitempty"`
	Owner          string                            `json:"owner,omitempty" yaml:"owner,omitempty"`
	VersionBundles []ClusterGuestConfigVersionBundle `json:"versionBundles,omitempty" yaml:"versionBundles,omitempty"`
}

type ClusterGuestConfigVersionBundle struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
}
