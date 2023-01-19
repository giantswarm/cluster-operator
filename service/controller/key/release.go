package key

import releasev1alpha1 "github.com/giantswarm/release-operator/v3/api/v1alpha1"

func HasCilium(release releasev1alpha1.Release) bool {
	for _, app := range release.Spec.Apps {
		if app.Name == "cilium" {
			return true
		}
	}

	return false
}
