package app

import (
	"github.com/giantswarm/microerror"
)

var notMigratedError = &microerror.Error{
	Kind: "notMigratedError",
}

// isNotMigratedError asserts notMigratedError.
func isNotMigratedError(err error) bool {
	return microerror.Cause(err) == notMigratedError
}
