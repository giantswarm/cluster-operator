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

// ChartConfigDeployPatch is used to patch a chartconfig resource
// in order to deploy a chart from a new channel.
type ChartConfigDeployPatch struct {
	Spec     Spec     `json:"spec"`
	Metadata Metadata `json:"metadata"`
}
