package unittest

import (
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
)

const (
	DefaultClusterID = "8y5ck"
)

func DefaultCluster() infrastructurev1alpha2.AWSCluster {
	cr := infrastructurev1alpha2.AWSCluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:         DefaultClusterID,
				label.OperatorVersion: "3.1.1",
				label.Release:         "100.0.0",
			},
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha2.AWSClusterSpec{
			Cluster: infrastructurev1alpha2.AWSClusterSpecCluster{
				Description: "Test cluster for template rendering unit test.",
				DNS: infrastructurev1alpha2.AWSClusterSpecClusterDNS{
					Domain: "gauss.eu-central-1.aws.gigantic.io",
				},
			},
			Provider: infrastructurev1alpha2.AWSClusterSpecProvider{
				CredentialSecret: infrastructurev1alpha2.AWSClusterSpecProviderCredentialSecret{
					Name:      "default-credential-secret",
					Namespace: "default",
				},
				Master: infrastructurev1alpha2.AWSClusterSpecProviderMaster{
					AvailabilityZone: "eu-central-1b",
					InstanceType:     "m5.xlarge",
				},
				Region: "eu-central-1",
			},
		},
		Status: infrastructurev1alpha2.AWSClusterStatus{
			Cluster: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
					{
						LastTransitionTime: metav1.NewTime(time.Now().Add(-90 * time.Minute)),
						Condition:          "Creating",
					},
					{
						LastTransitionTime: metav1.NewTime(time.Now().Add(-60 * time.Minute)),
						Condition:          "Created",
					},
					{
						LastTransitionTime: metav1.NewTime(time.Now().Add(-15 * time.Minute)),
						Condition:          "Updating",
					},
				},
				ID:       "yolo1",
				Versions: nil,
			},
			Provider: infrastructurev1alpha2.AWSClusterStatusProvider{
				Network: infrastructurev1alpha2.AWSClusterStatusProviderNetwork{
					CIDR: "10.0.0.0/24",
				},
			},
		},
	}

	return cr
}
