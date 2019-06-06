package updater

type responseError struct {
	Code  string
	Error string
}

// Response is the return value of the service action.
type Response struct {
}

// DefaultResponse provides a default response object by best effort.
func DefaultResponse() *Response {
	return &Response{}
}
