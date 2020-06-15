package nodecount

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var tenantClusterNotInitializedError = &microerror.Error{
	Kind: "tentantClusterNotInitializedError",
}

// IsTenantClusterNotInitialized asserts tenantClusterNotInitializedError.
func IsTenantClusterNotInitialized(err error) bool {
	return microerror.Cause(err) == tenantClusterNotInitializedError
}
