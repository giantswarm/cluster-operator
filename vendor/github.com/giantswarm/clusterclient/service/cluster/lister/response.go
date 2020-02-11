package lister

import "time"

// Response is the return value of the service action.
type Response struct {
	CreateDate time.Time `json:"create_date"`
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Release    string    `json:"release"`
}

// DefaultResponse provides a default response by best effort.
func DefaultResponse() []*Response {
	return []*Response{}
}
