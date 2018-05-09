// +build k8srequired

package setup

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/integration/teardown"
	"github.com/giantswarm/cluster-operator/integration/template"
	"github.com/giantswarm/cluster-operator/service"
)

func hostPeerVPC(c *awsclient.Client) error {
	log.Printf("Creating Host Peer VPC stack")

	clusterID := os.Getenv("CLUSTER_NAME")
	os.Setenv("AWS_ROUTE_TABLE_0", clusterID+"_0")
	os.Setenv("AWS_ROUTE_TABLE_1", clusterID+"_1")
	stackName := "host-peer-" + clusterID
	stackInput := &cloudformation.CreateStackInput{
		StackName:        aws.String(stackName),
		TemplateBody:     aws.String(os.ExpandEnv(e2etemplates.AWSHostVPCStack)),
		TimeoutInMinutes: aws.Int64(2),
	}
	_, err := c.CloudFormation.CreateStack(stackInput)
	if err != nil {
		return microerror.Mask(err)
	}
	err = c.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	describeOutput, err := c.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	for _, o := range describeOutput.Stacks[0].Outputs {
		if *o.OutputKey == "VPCID" {
			os.Setenv("AWS_VPC_PEER_ID", *o.OutputValue)
			break
		}
	}
	log.Printf("Host Peer VPC stack created")
	return nil
}

func WrapTestMain(g *framework.Guest, h *framework.Host, helmClient *helmclient.Client, apprClient *apprclient.Client, m *testing.M) {
	var v int
	var err error

	c := awsclient.NewClient()

	clusterName := fmt.Sprintf("ci-clop-%s-%s", os.Getenv("TESTED_VERSION"), os.Getenv("CIRCLE_SHA1")[0:5])
	os.Setenv("CLUSTER_NAME", clusterName)

	err = hostPeerVPC(c)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = h.Setup()
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	vbv, err := framework.GetVersionBundleVersion(service.NewVersionBundles(), os.Getenv("TESTED_VERSION"))
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}
	os.Setenv("CLOP_VERSION_BUNDLE_VERSION", vbv)

	err = resources(h, g, helmClient)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	if v == 0 {
		err = g.Initialize()
		if err != nil {
			log.Printf("%#v\n", err)
			v = 1
		}
		err = g.WaitForAPIUp()
		if err != nil {
			log.Printf("%#v\n", err)
			v = 1
		}
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
		name := "aws-operator"
		customResource := "awsconfig"
		logEntry := "deleted the guest cluster main stack"
		h.DeleteGuestCluster(name, customResource, logEntry)

		// only do full teardown when not on CI
		if os.Getenv("CIRCLECI") != "true" {
			err := teardown.Teardown(c, h, helmClient)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			h.Teardown()
		}
	}

	os.Exit(v)
}

func resources(h *framework.Host, g *framework.Guest, helmClient *helmclient.Client) error {
	err := h.InstallStableOperator("cert-operator", "certconfig", e2etemplates.CertOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}
	err = h.InstallStableOperator("node-operator", "nodeconfig", e2etemplates.NodeOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}
	err = h.InstallStableOperator("aws-operator", "awsconfig", e2etemplates.AWSOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}
	err = h.InstallCertResource()
	if err != nil {
		return microerror.Mask(err)
	}

	err = h.InstallResource("aws-resource-lab", e2etemplates.AWSResourceChartValues, ":stable")
	if err != nil {
		return microerror.Mask(err)
	}

	err = h.InstallBranchOperator("cluster-operator", "awsclusterconfig", template.ClusterOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}

	err = h.InstallResource("cluster-operator-resource", template.ClusterOperatorResourceChartValues, ":stable")
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
