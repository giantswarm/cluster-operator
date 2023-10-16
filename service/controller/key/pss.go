package key

import (
	"github.com/blang/semver"
	"github.com/giantswarm/microerror"
)

func IsPSSRelease(getter LabelsGetter) (bool, error) {
	release := ReleaseVersion(getter)

	if release == "" {
		// Unable to get release from CR.
		return false, microerror.Maskf(unknownReleaseError, "cannot determine release version from CR")
	}

	releaseVersion, err := semver.ParseTolerant(release)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return releaseVersion.Major >= 19 && releaseVersion.Minor >= 3, nil
}
