package awsconfig

import (
	"errors"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

var (
	// Error to return when simulating unknown error returned from Kubernetes
	// API client.
	unknownAPIError = errors.New("Unknown error from k8s API")
)

func alwaysReturnErrorReactor(err error) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "*",
		Resource: "*",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, err
		},
	}
}

func newAWSClusterConfig(clusterID, name string) *v1alpha1.AWSClusterConfig {
	return &v1alpha1.AWSClusterConfig{
		TypeMeta: v1.TypeMeta{
			Kind: "AWSClusterConfig",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.AWSClusterConfigSpec{
			Guest: v1alpha1.AWSClusterConfigSpecGuest{
				ClusterGuestConfig: v1alpha1.ClusterGuestConfig{
					ID:   clusterID,
					Name: name,
				},
			},
		},
	}
}

func newAWSConfig(clusterID string) *providerv1alpha1.AWSConfig {
	return &providerv1alpha1.AWSConfig{
		TypeMeta: v1.TypeMeta{
			Kind: "AWSConfig",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: clusterID,
		},
		Spec: providerv1alpha1.AWSConfigSpec{
			Cluster: providerv1alpha1.Cluster{
				ID: clusterID,
			},
		},
	}
}
