package certconfigcleansing

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r Resource) EnsureCreated(ctx context.Context, obj interface{}) error {

	currentMap := map[string]bool{}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Collecting all AWSClusterConfigs")

	rs, err := r.g8sClient.CoreV1alpha1().AWSClusterConfigs("").List(v1.ListOptions{})
	if err != nil {
		r.logger.LogCtx(ctx, "level", "error", "message", "could not get AWSClusterConfig resource", "stack", fmt.Sprintf("%#v", err))
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Collected %d AWSClusterConfigs", len(rs.Items)))

	for _, awsConfig := range rs.Items {
		currentMap[awsConfig.Spec.Guest.ID] = true
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Collecting all CertConfigs")

	certs, err := r.g8sClient.CoreV1alpha1().CertConfigs("").List(v1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Collected %d CertConfigs", len(certs.Items)))

	count := 0
	for _, cert := range certs.Items {

		clusterID := cert.Spec.Cert.ClusterID

		if _, ok := currentMap[clusterID]; !ok {

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("CertConfig %#q do not have related AWSClusterConfigs, going to delete it", cert.Name))

			err := r.g8sClient.CoreV1alpha1().CertConfigs(cert.Namespace).Delete(cert.Name, &v1.DeleteOptions{})
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("Failed to delete CertConfig %#q", cert.Name), "stack", fmt.Sprintf("%#v", err))
			} else {
				count++
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Successfully delete CertConfig %#q", cert.Name))
			}

		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("CertConfig %#q HAVE related AWSClusterConfigs, should not delete it", cert.Name))
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Total deleted %d certconfigs", count))

	return nil
}
