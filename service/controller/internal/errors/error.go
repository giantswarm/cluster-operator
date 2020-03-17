package errors

import (
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	chartConfigNotInstalledErrorText = "the server could not find the requested resource (get chartconfigs.core.giantswarm.io)"
)

var (
	chartConfigNotAvailablePatterns = []*regexp.Regexp{
		regexp.MustCompile(`[Get|Patch|Post] https://api\..*/apis/core.giantswarm.io/v1alpha1/namespaces.* (unexpected )?EOF`),
		regexp.MustCompile(`[Get|Patch|Post] https://api\..*/apis/core.giantswarm.io/v1alpha1/namespaces.* net/http: (TLS handshake timeout|request canceled).*?`),
	}
)

// ChartConfigNotAvailableError is returned when the chartconfig custom
// resources are not available in the tenant cluster.
var ChartConfigNotAvailableError = &microerror.Error{
	Kind: "ChartConfigNotAvailableError",
}

// ChartConfigNotInstalledError is returned when the chartconfig CRD is not
// installed in a tenant cluster.
var ChartConfigNotInstalledError = &microerror.Error{
	Kind: "ChartConfigNotInstalledError",
}

// IsChartConfigNotAvailable asserts ChartConfigNotAvailableError.
func IsChartConfigNotAvailable(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	for _, re := range chartConfigNotAvailablePatterns {
		matched := re.MatchString(c.Error())

		if matched {
			return true
		}
	}

	if c == ChartConfigNotAvailableError {
		return true
	}

	return false
}

// IsChartConfigNotInstalled asserts ChartConfigNotInstalledError
func IsChartConfigNotInstalled(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if strings.Contains(c.Error(), chartConfigNotInstalledErrorText) {
		return true
	}

	if c == ChartConfigNotInstalledError {
		return true
	}

	return false
}