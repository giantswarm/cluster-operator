package vault

import (
	"github.com/giantswarm/cluster-operator/flag/guest/cluster/vault/certificate"
)

// Vault is a data structure to hold guest cluster vault related configuration.
type Vault struct {
	Certificate certificate.Certificate
}
