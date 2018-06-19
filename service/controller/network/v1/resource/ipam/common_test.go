package ipam

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

var (
	// unknownAPIError for simulating unknown error returned from Kubernetes
	// API client.
	unknownAPIError = errors.New("Unknown error from k8s API")
)

type k8stestingReactorFactoryFunc func(t *testing.T, expectedStatus *v1alpha1.ClusterNetworkConfigStatus) k8stesting.Reactor

func mustParseNetworkCIDR(cidr string) net.IPNet {
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	return *n
}

func alwaysFailReactor(t *testing.T, _ *v1alpha1.ClusterNetworkConfigStatus) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "*",
		Resource: "*",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			t.Fatal("Kubernetes API invoked when logic must not.")
			return true, nil, nil
		},
	}
}

func alwaysReturnErrorReactor(err error) k8stestingReactorFactoryFunc {
	return func(t *testing.T, _ *v1alpha1.ClusterNetworkConfigStatus) k8stesting.Reactor {
		return &k8stesting.SimpleReactor{
			Verb:     "*",
			Resource: "*",
			Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, err
			},
		}
	}
}

func verifyClusterNetworkConfigStatusUpdateReactor(t *testing.T, expectedStatus *v1alpha1.ClusterNetworkConfigStatus) k8stesting.Reactor {
	return &k8stesting.SimpleReactor{
		Verb:     "update",
		Resource: "clusternetworkconfigs",
		Reaction: func(action k8stesting.Action) (bool, runtime.Object, error) {
			updateAction, ok := action.(k8stesting.UpdateActionImpl)
			if !ok {
				return false, nil, microerror.Maskf(wrongTypeError, "action != k8stesting.UpdateActionImpl")
			}

			var updatedClusterNetworkConfig *v1alpha1.ClusterNetworkConfig
			if updateActionObj := updateAction.GetObject(); updateActionObj != nil {
				updatedClusterNetworkConfig, ok = updateActionObj.(*v1alpha1.ClusterNetworkConfig)
				if !ok {
					return false, nil, microerror.Maskf(wrongTypeError, "UpdateAction did not contain *v1alpha1.ClusterNetworkConfigStatus")
				}
			}

			if expectedStatus == nil {
				t.Fatalf("verifyClusterNetworkConfigStatusUpdateReactor() called with expectedStatus == %v", expectedStatus)
			}

			if !reflect.DeepEqual(*expectedStatus, updatedClusterNetworkConfig.Status) {
				t.Fatalf("UpdateStatus() got %#v, expected %#v", updatedClusterNetworkConfig.Status, expectedStatus)
			}

			return true, updatedClusterNetworkConfig, nil
		},
	}
}
