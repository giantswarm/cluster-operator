package lister

import (
	"time"

	"github.com/giantswarm/versionbundle"
)

// Response is the return value of the service action.
type Response struct {
	Active     bool                `json:"active"`
	Apps       []versionbundle.App `json:"apps"`
	Changelogs []Changelog         `json:"changelogs"`
	Components []Component         `json:"components"`
	Timestamp  time.Time           `json:"timestamp"`
	Version    string              `json:"version"`
}

// DefaultResponse provides a default response by best effort.
func DefaultResponse() []Response {
	return []Response{}
}
