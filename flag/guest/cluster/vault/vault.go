package vault

import (
	"github.com/giantswarm/kubernetesd/flag/service/vault/certificate"
)

// Vault is a data structure to hold guest cluster vault related configuration.
type Vault struct {
	Certificate certificate.Certificate
}
