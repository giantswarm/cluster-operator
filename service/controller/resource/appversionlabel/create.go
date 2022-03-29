package appversionlabel

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/apiextensions/v6/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/cluster-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/internal/releaseversion"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var apps []*v1alpha1.App
	{
		r.logger.Debugf(ctx, "finding optional apps for tenant cluster %#q", key.ClusterID(&cr))

		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s,%s!=%s", label.Cluster, key.ClusterID(&cr), label.ManagedBy, project.Name()),
		}

		list := &v1alpha1.AppList{}
		err := r.ctrlClient.List(ctx, list, &client.ListOptions{Raw: &o})
		if err != nil {
			return microerror.Mask(err)
		}

		for _, item := range list.Items {
			apps = append(apps, item.DeepCopy())
		}

		r.logger.Debugf(ctx, "found %d optional apps for tenant cluster %#q", len(apps), key.ClusterID(&cr))
	}

	{
		var updatedAppCount int

		if len(apps) > 0 {
			componentVersions, err := r.releaseVersion.ComponentVersion(ctx, &cr)
			if err != nil {
				return microerror.Mask(err)
			}

			appOperatorComponent := componentVersions[releaseversion.AppOperator]
			appOperatorVersion := appOperatorComponent.Version
			if appOperatorVersion == "" {
				return microerror.Maskf(notFoundError, "app-operator component version not found")
			}

			r.logger.Debugf(ctx, "updating version label for optional apps in tenant cluster %#q", key.ClusterID(&cr))

			for _, app := range apps {
				currentVersion := app.Labels[label.AppOperatorVersion]

				if currentVersion != appOperatorVersion {
					patches := []patch{}

					if len(app.Labels) == 0 {
						patches = append(patches, patch{
							Op:    "add",
							Path:  "/metadata/labels",
							Value: map[string]string{},
						})
					}

					patches = append(patches, patch{
						Op:    "add",
						Path:  fmt.Sprintf("/metadata/labels/%s", replaceToEscape(label.AppOperatorVersion)),
						Value: appOperatorVersion,
					})

					bytes, err := json.Marshal(patches)
					if err != nil {
						return microerror.Mask(err)
					}

					err = r.ctrlClient.Patch(ctx, app, client.RawPatch(types.JSONPatchType, bytes), &client.PatchOptions{Raw: &metav1.PatchOptions{}})
					if err != nil {
						return microerror.Mask(err)
					}

					updatedAppCount++
				}
			}

			r.logger.Debugf(ctx, "updating version label for %d optional apps in tenant cluster %#q", updatedAppCount, key.ClusterID(&cr))
		}
	}

	return nil
}
