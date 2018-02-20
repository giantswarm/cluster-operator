package certconfig

import "github.com/giantswarm/microerror"

// KeyError is an interface that defines contract between Key error
// check implementations in CRD frameworks and certconfig resource.
type KeyError interface {
	// IsWrongTypeError asserts if error is caused by type mismatch.
	IsWrongTypeError(err error) bool
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
