package key

import (
	"encoding/json"
	"fmt"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/runtime"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func clusterProviderSpec(cluster cmav1alpha1.Cluster) g8sv1alpha1.AWSClusterSpec {
	spec, err := g8sClusterSpecFromCMAClusterSpec(cluster.Spec.ProviderSpec)
	if err != nil {
		panic(err)
	}
	return spec
}

func clusterProviderStatus(cluster cmav1alpha1.Cluster) g8sv1alpha1.AWSClusterStatus {
	return g8sClusterStatusFromCMAClusterStatus(cluster.Status.ProviderStatus)
}

func g8sClusterSpecFromCMAClusterSpec(cmaSpec cmav1alpha1.ProviderSpec) (g8sv1alpha1.AWSClusterSpec, error) {
	if cmaSpec.Value == nil {
		return g8sv1alpha1.AWSClusterSpec{}, microerror.Maskf(notFoundError, "provider spec extension for AWS not found")
	}

	var g8sSpec g8sv1alpha1.AWSClusterSpec
	{
		if len(cmaSpec.Value.Raw) == 0 {
			return g8sSpec, nil
		}

		err := json.Unmarshal(cmaSpec.Value.Raw, &g8sSpec)
		if err != nil {
			return g8sv1alpha1.AWSClusterSpec{}, microerror.Mask(err)
		}
	}

	// Debug; will be removed in a bit.
	fmt.Printf("\n\n\n\n\n%#v\n\n\n\n\n\n", g8sSpec)

	return g8sSpec, nil
}

func g8sClusterStatusFromCMAClusterStatus(cmaStatus *runtime.RawExtension) g8sv1alpha1.AWSClusterStatus {
	var g8sStatus g8sv1alpha1.AWSClusterStatus
	{
		if cmaStatus == nil {
			return g8sStatus
		}

		if len(cmaStatus.Raw) == 0 {
			return g8sStatus
		}

		err := json.Unmarshal(cmaStatus.Raw, &g8sStatus)
		if err != nil {
			panic(err)
		}
	}

	return g8sStatus
}
