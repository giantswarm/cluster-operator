package guestcluster

import (
	"regexp"

	"github.com/giantswarm/microerror"
)

const (
	// guestDNSNotReadyPattern is a regular expression representing DNS errors for
	// the guest API domain.
	// match example https://play.golang.org/p/ipBkwqlc4Td
	guestDNSNotReadyPattern = "dial tcp: lookup .* on .*:53: no such host"

	// guestTransientInvalidCertificatePattern regular expression defines the kind
	// of transient errors related to certificates returned while the guest API is
	// not fully up.
	// match example https://play.golang.org/p/iiYvBhPOg4f
	guestTransientInvalidCertificatePattern = `[Get|Post] https://api\..*: x509: certificate is valid for ingress.local, not api\..*`
)

var (
	guestDNSNotReadyRegexp                 *regexp.Regexp
	guestTransientInvalidCertificateRegexp *regexp.Regexp
)

func init() {
	guestDNSNotReadyRegexp = regexp.MustCompile(guestDNSNotReadyPattern)
	guestTransientInvalidCertificateRegexp = regexp.MustCompile(guestTransientInvalidCertificatePattern)
}

var guestAPINotAvailableError = &microerror.Error{
	Kind: "guestAPINotAvailableError",
}

// IsGuestAPINotAvailable asserts guestAPINotAvailableError.
func IsGuestAPINotAvailable(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	regexps := []*regexp.Regexp{
		guestDNSNotReadyRegexp,
		guestTransientInvalidCertificateRegexp,
	}
	for _, re := range regexps {
		matched := re.MatchString(c.Error())

		if matched {
			return true
		}
	}

	if c == guestAPINotAvailableError {
		return true
	}

	return false
}

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
