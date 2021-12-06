package podcidr

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidTypeError = &microerror.Error{
	Kind: "invalidTypeError",
}

// IsInvalidType asserts invalidTypeError.
func IsInvalidType(err error) bool {
	return microerror.Cause(err) == invalidTypeError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var tooManyCRsError = &microerror.Error{
	Kind: "tooManyCRsError",
	Desc: "There is only a single AWSCluster CR allowed with the current implementation.",
}

// IsTooManyCRsError asserts tooManyCRsError.
func IsTooManyCRsError(err error) bool {
	return microerror.Cause(err) == tooManyCRsError
}

var unsupportedProviderError = &microerror.Error{
	Kind: "unsupportedProviderError",
}

// IsUnsupportedProvider asserts unsupportedProviderError.
func IsUnsupportedProvider(err error) bool {
	return microerror.Cause(err) == unsupportedProviderError
}
