package appversionlabel

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var apps []*v1alpha1.App
	{
		_ = r.logger.LogCtx(ctx, fmt.Sprintf("finding optional apps for tenant cluster %#q", clusterConfig.ID))

		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s!=%s", label.ManagedBy, project.Name()),
		}
		list, err := r.g8sClient.ApplicationV1alpha1().Apps(clusterConfig.ID).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, item := range list.Items {
			apps = append(apps, item.DeepCopy())
		}

		_ = r.logger.LogCtx(ctx, fmt.Sprintf("found %d optional apps for tenant cluster %#q", len(apps), clusterConfig.ID))
	}

	{
		var updatedAppCount int

		if len(apps) > 0 {
			// put `v` as a prefix of release version since all releases CRs keep this format.
			appOperatorVersion, err := r.getComponentVersion(fmt.Sprintf("v%s", clusterConfig.ReleaseVersion), "app-operator")
			if err != nil {
				return microerror.Mask(err)
			}

			_ = r.logger.LogCtx(ctx, fmt.Sprintf("updating version label for optional apps in tenant cluster %#q", clusterConfig.ID))

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

					_, err = r.g8sClient.ApplicationV1alpha1().Apps(app.Namespace).Patch(app.Name, types.JSONPatchType, bytes)
					if err != nil {
						return microerror.Mask(err)
					}

					updatedAppCount++
				}
			}

			r.logger.LogCtx(ctx, fmt.Sprintf("updating version label for %d optional apps in tenant cluster %#q", updatedAppCount, clusterConfig.ID))
		}
	}

	return nil
}
