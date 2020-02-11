package app

import (
	"github.com/giantswarm/microerror"
)

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfigError asserts invalidConfigError.
func IsInvalidConfigError(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notMigratedError = &microerror.Error{
	Kind: "notMigrateError",
}

// IsNotMigratedError asserts notMigratedError.
func IsNotMigratedError(err error) bool {
	return microerror.Cause(err) == notMigratedError
}
