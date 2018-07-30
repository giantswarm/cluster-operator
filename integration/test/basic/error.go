package basic

import (
	"github.com/giantswarm/microerror"
)

var emptyChartConfigListError = &microerror.Error{
	Kind: "emptyChartConfigListError",
}

// IsEmptyChartConfigList asserts emptyChartConfigListError.
func IsEmptyChartConfigList(err error) bool {
	return microerror.Cause(err) == emptyChartConfigListError
}

var releaseStatusNotMatchingError = &microerror.Error{
	Kind: "releaseStatusNotMatchingError",
}

// IsReleaseStatusNotMatching asserts releaseStatusNotMatchingError
func IsReleaseStatusNotMatching(err error) bool {
	return microerror.Cause(err) == releaseStatusNotMatchingError
}
