package unittest

import (
	"time"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultRelease() releasev1alpha1.Release {
	r := releasev1alpha1.Release{
		ObjectMeta: metav1.ObjectMeta{
			Name: "v11.0.1",
		},
		Spec: releasev1alpha1.ReleaseSpec{
			Apps: []releasev1alpha1.ReleaseSpecApp{
				{
					Name:    "cert-operator",
					Version: "1.2.1",
				},
				{
					Name:    "chart-operator",
					Version: "0.11.4",
				},
				{
					ComponentVersion: "1.6.5",
					Name:             "coredns",
					Version:          "1.1.3",
				},
			},
			Components: []releasev1alpha1.ReleaseSpecComponent{
				{
					Name:    "app-operator",
					Version: "1.0.0",
				},
				{
					Name:    "cert-operator",
					Version: "1.0.0",
				},
				{
					Name:    "cluster-operator",
					Version: "0.23.1",
				},
			},
			Date: &metav1.Time{
				Time: time.Now(),
			},
			State: "active",
		},
	}

	return r
}
