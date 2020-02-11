package creator

import "time"

// Response is the return value of the service action.
type Response struct {
	CertificateAuthorityData string    `json:"certificate_authority_data"`
	ClientCertificateData    string    `json:"client_certificate_data"`
	ClientKeyData            string    `json:"client_key_data"`
	CreateDate               time.Time `json:"create_date"`
	Description              string    `json:"description"`
	SerialNumber             string    `json:"serial_number"`
	TTL                      int       `json:"ttl"`
}

// DefaultResponse provides a default response by best effort.
func DefaultResponse() *Response {
	return &Response{
		CertificateAuthorityData: "",
		ClientCertificateData:    "",
		ClientKeyData:            "",
		CreateDate:               time.Time{},
		Description:              "",
		SerialNumber:             "",
		TTL:                      0,
	}
}
