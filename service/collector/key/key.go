package key

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
)

func AWSAPIDomain(cr v1alpha1.AWSClusterConfig) string {
	return fmt.Sprintf("api.%s", cr.Spec.Guest.DNSZone)
}

func AWSClusterID(cr v1alpha1.AWSClusterConfig) string {
	return cr.Spec.Guest.ID
}

func AzureAPIDomain(cr v1alpha1.AzureClusterConfig) string {
	return fmt.Sprintf("api.%s", cr.Spec.Guest.DNSZone)
}

func AzureClusterID(cr v1alpha1.AzureClusterConfig) string {
	return cr.Spec.Guest.ID
}

func ChartOperatorReleaseName() string {
	return "chart-operator"
}

func KVMAPIDomain(cr v1alpha1.KVMClusterConfig) string {
	return fmt.Sprintf("api.%s", cr.Spec.Guest.DNSZone)
}

func KVMClusterID(cr v1alpha1.KVMClusterConfig) string {
	return cr.Spec.Guest.ID
}
