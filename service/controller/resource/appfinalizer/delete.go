package appfinalizer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

// EnsureDeleted removes finalizers for workload cluster app CRs. These are
// deleted with the cluster by the provider operator.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// We keep the finalizer for the app-operator app CR so the resources in
	// the management cluster are deleted.
	o := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s!=%s", label.AppKubernetesName, "app-operator"),
	}

	r.logger.Debugf(ctx, "finding apps to remove finalizers for")

	list, err := r.g8sClient.ApplicationV1alpha1().Apps(key.ClusterID(&cr)).List(ctx, o)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "found %d apps to remove finalizers for", len(list.Items))

	for _, app := range list.Items {
		r.logger.Debugf(ctx, "removing finalizer for app %#q", app.Name)

		index := getFinalizerIndex(app.Finalizers)
		if index >= 0 {
			patches := []patch{
				{
					Op:   "remove",
					Path: fmt.Sprintf("/metadata/finalizers/%d", index),
				},
			}
			bytes, err := json.Marshal(patches)
			if err != nil {
				return microerror.Mask(err)
			}

			_, err = r.g8sClient.ApplicationV1alpha1().Apps(app.Namespace).Patch(ctx, app.Name, types.JSONPatchType, bytes, metav1.PatchOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "removed finalizer for app %#q", app.Name)
		} else {
			r.logger.Debugf(ctx, "finalizer already removed for app %#q", app.Name)
		}
	}

	return nil
}

func getFinalizerIndex(finalizers []string) int {
	for i, f := range finalizers {
		if f == "operatorkit.giantswarm.io/app-operator-app" {
			return i
		}
	}

	return -1
}
