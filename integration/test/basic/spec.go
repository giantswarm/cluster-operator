// +build k8srequired

package basic

type Chart struct {
	Channel string `json:"channel"`
}

type Metadata struct {
	Labels map[string]string `json:"labels"`
}

type Spec struct {
	Chart Chart `json:"chart"`
}

type ChartConfigDeployPatch struct {
	Spec     Spec     `json:"spec"`
	Metadata Metadata `json:"metadata"`
}
