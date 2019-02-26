package chartoperator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/tenantcluster"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/cluster-operator/pkg/v12/key"
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
			} else if helmclient.IsTillerInstallationFailed(err) {
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
				Tiller: Tiller{
					Namespace: chartOperatorNamespace,
				},
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

		var tenantK8sClient kubernetes.Interface
		{
			tenantK8sClient, err = r.getTenantK8sClient(ctx, obj)
			if tenantcluster.IsTimeout(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")

				// A timeout error here means that the cluster-operator certificate
				// for the current guest cluster was not found. We can't continue
				// without a Helm client. We will retry during the next execution, when
				// the certificate might be available.
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
			// We wait for the chart-operator deployment to be ready so the
			// chartconfig CRD is installed. This allows the chartconfig
			// resource to create CRs in the same reconcilation loop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for ready chart-operator deployment")

			o := func() error {
				err := r.checkDeploymentReady(ctx, tenantK8sClient, chartOperatorNamespace, chartOperatorDeployment)
				if err != nil {
					return microerror.Mask(err)
				}

				return nil
			}

			// Wait for chart-operator to be deployed. If it takes longer than
			// the timeout the chartconfig CRs will be created during the next
			// reconciliation loop.
			b := backoff.NewConstant(30*time.Second, 5*time.Second)
			n := func(err error, delay time.Duration) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%#q deployment is not ready retrying in %s", chartOperatorDeployment, delay), "stack", fmt.Sprintf("%#v", err))
			}

			err = backoff.RetryNotify(o, b, n)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "chart-operator deployment is ready")
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

// checkDeploymentReady checks for the specified deployment that the number of
// ready replicas matches the desired state.
func (r *Resource) checkDeploymentReady(ctx context.Context, k8sClient kubernetes.Interface, namespace, deploymentName string) error {
	deploy, err := k8sClient.Extensions().Deployments(namespace).Get(deploymentName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return microerror.Maskf(notReadyError, "deployment %#q not found", deploymentName)
	} else if err != nil {
		return microerror.Mask(err)
	}

	if deploy.Status.ReadyReplicas != *deploy.Spec.Replicas {
		return microerror.Maskf(notReadyError, "deployment %#q want %d replicas %d ready", deploymentName, *deploy.Spec.Replicas, deploy.Status.ReadyReplicas)
	}

	// Deployment is ready.
	return nil
}
