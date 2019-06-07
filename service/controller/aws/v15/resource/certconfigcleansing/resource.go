package certconfigcleansing

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "certconfigcleansingv15"
)

// Config represents the configuration used to create a new certconfigcleansing resource.
type Config struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

// Resource implements the certconfigcleansing resource.
type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func (r Resource) EnsureCreated(ctx context.Context, obj interface{}) error {

	currentMap := map[string]bool{}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Collecting all AWSClusterConfigs")

	rs, err := r.g8sClient.CoreV1alpha1().AWSClusterConfigs("").List(metav1.ListOptions{})
	if err != nil {
		r.logger.LogCtx(ctx, "level", "error", "message", "could not get AWSClusterConfig resource", "stack", fmt.Sprintf("%#v", err))
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Collected %d AWSClusterConfigs", len(rs.Items)))

	for _, awsConfig := range rs.Items {
		currentMap[awsConfig.Spec.Guest.ID] = true
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Collecting all CertConfigs")

	certs, err := r.g8sClient.CoreV1alpha1().CertConfigs("").List(metav1.ListOptions{})
	if err != nil {
		r.logger.LogCtx(ctx, "level", "error", "message", "could not get CertConfigs resource", "stack", fmt.Sprintf("%#v", err))
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Collected %d CertConfigs", len(certs.Items)))

	count := 0
	for _, cert := range certs.Items {

		clusterID := cert.Spec.Cert.ClusterID

		if _, ok := currentMap[clusterID]; !ok {

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("CertConfig %#q do not have related AWSClusterConfigs, going to delete it", cert.Name))

			err := r.g8sClient.CoreV1alpha1().CertConfigs("").Delete(cert.Name, &metav1.DeleteOptions{})
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

func (r Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r Resource) Name() string {
	return Name
}
