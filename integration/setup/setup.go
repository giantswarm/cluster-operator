// +build k8srequired

package setup

import (
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apprclient"
	awsclient "github.com/giantswarm/aws-operator/integration/client"
	awstemplate "github.com/giantswarm/aws-operator/integration/template"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-operator/integration/teardown"
	"github.com/giantswarm/cluster-operator/integration/template"
)

func hostPeerVPC(c *awsclient.AWS, g *framework.Guest, h *framework.Host) error {
	log.Printf("Creating Host Peer VPC stack")

	clusterID := os.Getenv("CLUSTER_NAME")
	os.Setenv("AWS_ROUTE_TABLE_0", clusterID+"_0")
	os.Setenv("AWS_ROUTE_TABLE_1", clusterID+"_1")
	stackName := "host-peer-" + clusterID
	stackInput := &cloudformation.CreateStackInput{
		StackName:        aws.String(stackName),
		TemplateBody:     aws.String(os.ExpandEnv(awstemplate.AWSHostVPCStack)),
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

	c := awsclient.NewAWS()

	err = hostPeerVPC(c, g, h)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = h.Setup()
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = resources(c, h, g, helmClient)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	if v == 0 {
		err = g.Setup()
		if err != nil {
			log.Printf("%#v\n", err)
			v = 1
		}
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
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

func resources(c *awsclient.AWS, h *framework.Host, g *framework.Guest, helmClient *helmclient.Client) error {
	err := h.InstallStableOperator("cert-operator", "certconfig", awstemplate.CertOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}
	err = h.InstallStableOperator("node-operator", "nodeconfig", awstemplate.NodeOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}
	err = h.InstallStableOperator("aws-operator", "awsconfig", awstemplate.AWSOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}
	err = h.InstallCertResource()
	if err != nil {
		return microerror.Mask(err)
	}
	// TODO this should probably be in the e2e-harness framework as well just like
	// the other stuff.
	err = h.InstallResource("aws-resource-lab", awstemplate.AWSResourceChartValues, ":stable")
	if err != nil {
		return microerror.Mask(err)
	}

	err = h.InstallBranchOperator("cluster-operator", "awsclusterconfig", template.ClusterOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
