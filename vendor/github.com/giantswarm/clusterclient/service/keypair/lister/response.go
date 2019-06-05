package lister

import "time"

// Response is the return value of the service action.
type Response struct {
	// CommonName is the common name that the certifcate was issued with.
	CommonName string `json:"common_name"`
	// CreateDate represents the timestamp of the serial items creation.
	CreateDate time.Time `json:"create_date"`
	// Description represents the description associated with an issued key-pair.
	Description string `json:"description"`
	// Organizations represents the organizations that the certificate was issued with.
	Organizations string `json:"certificate_organizations"`
	// SerialNumber represents the fingerprint of an issued key-pair.
	SerialNumber string `json:"serial_number"`
	// TTL represents the time the number of hours the key-pair associated with
	// the current serial item is valid.
	TTL int `json:"ttl"`
}

// DefaultResponse provides a default response by best effort.
func DefaultResponse() []*Response {
	return []*Response{}
}
