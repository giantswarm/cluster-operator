package chartoperator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/cluster-operator/pkg/v10/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	createState, err := toResourceState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if !reflect.DeepEqual(createState, ResourceState{}) {
		var tenantHelmClient helmclient.Interface
		{
			tenantHelmClient, err = r.getTenantHelmClient(ctx, obj)
			if tenantcluster.IsTimeout(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")

				// A timeout error here means that the cluster-operator certificate
				// for the current guest cluster was not found. We can't continue
				// without a Helm client. We will retry during the next execution, when
				// the certificate might be available.
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)

				return nil
			} else if helmclient.IsTillerNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "Tiller installation failed")

				// Tiller installation can fail during guest cluster setup. We will
				// retry on next reconciliation loop.
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)

				return nil
			} else if guest.IsAPINotAvailable(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "guest API not available")

				// We should not hammer guest API if it is not available, the guest
				// cluster might be initializing. We will retry on next reconciliation
				// loop.
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)

				return nil
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "creating chart-operator chart")

			p, err := r.apprClient.PullChartTarball(ctx, createState.ChartName, chartOperatorChannel)
			if err != nil {
				return microerror.Mask(err)
			}
			defer func() {
				err := r.fs.Remove(p)
				if err != nil {
					r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %q failed", p), "stack", fmt.Sprintf("%#v", err))
				}
			}()

			clusterDNSIP, err := key.DNSIP(r.clusterIPRange)
			if err != nil {
				return microerror.Mask(err)
			}
			v := &Values{
				ClusterDNSIP: clusterDNSIP,
				Image: Image{
					Registry: r.registryDomain,
				},
				TillerNamespace: chartOperatorNamespace,
			}
			b, err := json.Marshal(v)
			if err != nil {
				return microerror.Mask(err)
			}

			err = tenantHelmClient.InstallReleaseFromTarball(ctx, p, chartOperatorNamespace,
				helm.ReleaseName(createState.ReleaseName),
				helm.ValueOverrides(b),
				helm.InstallWait(true),
			)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "created chart-operator chart")
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not create chart-operator chart")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentResourceState, err := toResourceState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredResourceState, err := toResourceState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if chart-operator chart has to be created")

	createState := &ResourceState{}

	// chart-operator should be created if it is not present.
	if reflect.DeepEqual(currentResourceState, ResourceState{}) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chart-operator chart needs to be created")

		createState = &desiredResourceState
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chart-operator chart does not need to be created")
	}

	return createState, nil
}
