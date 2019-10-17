package appmigration

import (
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	errorText = "the server could not find the requested resource (get chartconfigs.core.giantswarm.io)"
)

func isChartConfigNotInstalled(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	return strings.Contains(c.Error(), errorText)
}
