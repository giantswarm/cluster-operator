package appversionlabel

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/apiextensions/v6/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/cluster-operator/v5/service/controller/key"
	"github.com/giantswarm/cluster-operator/v5/service/internal/releaseversion"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cluster, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	logger := r.logger.With("cluster", key.ClusterID(&cluster))

	apps, err := r.getApps(ctx, logger, cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.updateApps(ctx, logger, cluster, apps)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (r *Resource) getApps(ctx context.Context, logger micrologger.Logger, cluster apiv1beta1.Cluster) ([]*v1alpha1.App, error) {
	logger.Debugf(ctx, "finding apps")

	list := &v1alpha1.AppList{}
	err := r.ctrlClient.List(ctx, list, &client.ListOptions{Namespace: key.ClusterID(&cluster), Raw: &metav1.ListOptions{}})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	apps := []*v1alpha1.App{}
	for _, item := range list.Items {
		apps = append(apps, item.DeepCopy())
	}

	logger.Debugf(ctx, "found %d apps", len(apps))

	return apps, nil
}

func (r *Resource) updateApps(ctx context.Context, logger micrologger.Logger, cluster apiv1beta1.Cluster, apps []*v1alpha1.App) error {
	if len(apps) == 0 {
		return nil
	}

	appOperatorVersion, err := r.getAppOperatorVersion(ctx, cluster)
	if err != nil {
		return err
	}

	logger.Debugf(ctx, "updating app version labels")

	updatedAppCount := 0
	for _, app := range apps {
		currentVersion := app.Labels[label.AppOperatorVersion]

		if !shouldUpdateAppOperatorVersionLabel(currentVersion, appOperatorVersion) {
			continue
		}

		err = r.patchAppOperatorVersion(ctx, app, appOperatorVersion)
		if err != nil {
			return err
		}

		updatedAppCount++
	}

	logger.Debugf(ctx, "updated version label for %d apps", updatedAppCount)

	return nil
}

func (r *Resource) getAppOperatorVersion(ctx context.Context, cluster apiv1beta1.Cluster) (string, error) {
	componentVersions, err := r.releaseVersion.ComponentVersion(ctx, &cluster)
	if err != nil {
		return "", microerror.Mask(err)
	}

	appOperatorComponent := componentVersions[releaseversion.AppOperator]
	appOperatorVersion := appOperatorComponent.Version
	if appOperatorVersion == "" {
		return "", microerror.Maskf(notFoundError, "app-operator component version not found")
	}

	return appOperatorVersion, nil
}

func (r *Resource) patchAppOperatorVersion(ctx context.Context, app *v1alpha1.App, appOperatorVersion string) error {
	var patches []patch

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

	err = r.ctrlClient.Patch(ctx,
		app,
		client.RawPatch(types.JSONPatchType, bytes),
		&client.PatchOptions{Raw: &metav1.PatchOptions{}},
	)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// shouldUpdateAppOperatorVersionLabel When the current version is 0.0.0  aka they are reconciled by the management
// cluster app-operator. This is a use-case for App Bundles  for example, because the App CRs they contain should be
// created in the management cluster so should be reconciled by the management cluster app-operator.
func shouldUpdateAppOperatorVersionLabel(currentVersion string, componentVersion string) bool {
	if currentVersion == key.UniqueOperatorVersion {
		return false
	}

	return currentVersion != componentVersion
}
