// +build k8srequired

package teardown

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"k8s.io/helm/pkg/helm"
)

func Resources(ctx context.Context, c *awsclient.Client, f *framework.Host, helmClient *helmclient.Client) error {
	errors := make([]error, 0)

	targetNamespace := "default"
	items := []string{
		"cluster-operator",
		"apiextensions-aws-cluster-config-e2e",
		"cert-operator",
		"cert-config-e2e",
		"aws-operator",
		"apiextensions-aws-config-e2e",
		"node-operator",
	}

	for _, item := range items {
		releaseName := fmt.Sprintf("%s-%s", targetNamespace, item)
		log.Printf("deleting release %#q", releaseName)

		err := helmClient.DeleteRelease(ctx, releaseName, helm.DeletePurge(true))
		if err != nil {
			log.Printf("failed to delete release %#q %#v", releaseName, err)
			errors = append(errors, err)
		} else {
			log.Printf("deleted release %#q", releaseName)
		}
	}

	if len(errors) != 0 {
		return microerror.Mask(errors[0])
	}

	return nil
}

func HostPeerVPC(c *awsclient.Client) error {
	log.Printf("Deleting Host Peer VPC stack")

	_, err := c.CloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String("host-peer-" + os.Getenv("CLUSTER_NAME")),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
