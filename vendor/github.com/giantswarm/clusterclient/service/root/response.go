package root

// Response is the return value of the service action.
type Response struct {
	Description    string `json:"description"`
	GitCommit      string `json:"git-commit"`
	GoVersion      string `json:"go-version"`
	Name           string `json:"name"`
	OSArch         string `json:"os-arch"`
	ProjectVersion string `json:"project-version"`
	Source         string `json:"source"`
}

// DefaultResponse provides a default response object by best effort.
func DefaultResponse() *Response {
	return &Response{
		Description:    "",
		GitCommit:      "",
		GoVersion:      "",
		Name:           "",
		OSArch:         "",
		ProjectVersion: "",
		Source:         "",
	}
}
