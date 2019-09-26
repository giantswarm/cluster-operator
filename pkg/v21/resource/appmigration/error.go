package appmigration

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfigError asserts invalidConfigError.
func IsInvalidConfigError(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notDeployedError = &microerror.Error{
	Kind: "notDeployedError",
}

// IsNotDeployed asserts notDeployedError.
func IsNotDeployed(err error) bool {
	return microerror.Cause(err) == notDeployedError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
