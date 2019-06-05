package lister

import "time"

// Response is the return value of the service action.
type Response struct {
	Active     bool        `json:"active"`
	Changelogs []Changelog `json:"changelogs"`
	Components []Component `json:"components"`
	Timestamp  time.Time   `json:"timestamp"`
	Version    string      `json:"version"`
}

// DefaultResponse provides a default response by best effort.
func DefaultResponse() []Response {
	return []Response{}
}
