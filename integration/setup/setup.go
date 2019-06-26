// +build k8srequired

package setup

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/e2esetup/k8s"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/integration/env"
	"github.com/giantswarm/cluster-operator/integration/teardown"
)

const (
	awsOperatorArnKey   = "aws.awsoperator.arn"
	credentialName      = "credential-default"
	credentialNamespace = "giantswarm"
)

const (
	operatorNamespace = "giantswarm"
)

func hostPeerVPC(c *awsclient.Client) error {
	log.Printf("Creating Host Peer VPC stack")

	os.Setenv("AWS_ROUTE_TABLE_0", env.ClusterID()+"_0")
	os.Setenv("AWS_ROUTE_TABLE_1", env.ClusterID()+"_1")
	stackName := "host-peer-" + env.ClusterID()
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

func WrapTestMain(ctx context.Context, g *framework.Guest, h *framework.Host, s *k8s.Setup, helmClient *helmclient.Client, apprClient *apprclient.Client, m *testing.M) {
	var v int
	var err error

	c, err := awsclient.NewClient()
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
		return
	}

	defer func() {
		if os.Getenv("KEEP_RESOURCES") != "true" {

			h.DeleteGuestCluster(ctx, "aws")

			err := teardown.HostPeerVPC(c)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}

			// only do full teardown when not on CI
			if os.Getenv("CIRCLECI") != "true" {
				err := teardown.Resources(ctx, c, h, helmClient)
				if err != nil {
					log.Printf("%#v\n", err)
					v = 1
				}
				// TODO there should be error handling for the framework teardown.
				h.Teardown()
			}
		}
		os.Exit(v)
	}()

	token := os.Getenv("GITHUB_BOT_TOKEN")
	vType := os.Getenv("TESTED_VERSION")
	params := &framework.VBVParams{
		Provider: "aws",
		Token:    token,
		VType:    vType,
	}
	authorities, err := framework.GetAuthorities(params)
	// do not fail on missing WIP.
	if os.Getenv("TESTED_VERSION") == "wip" && framework.IsNotFound(err) {
		log.Printf("WIP version not present, exiting.\n")
		os.Exit(0)
	}
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
		return
	}
	for _, authority := range authorities {
		switch authority.Name {
		case "aws-operator":
			os.Setenv("AWSOP_VERSION_BUNDLE_VERSION", authority.Version)
			// next env var is required by aws-operator templates.
			os.Setenv("VERSION_BUNDLE_VERSION", authority.Version)
		case "cluster-operator":
			os.Setenv("CLOP_VERSION_BUNDLE_VERSION", authority.Version)
		case "cert-operator":
			os.Setenv("CERTOP_VERSION_BUNDLE_VERSION", authority.Version)
		}
	}

	err = hostPeerVPC(c)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
		return
	}

	err = resources(ctx, h, g, s, helmClient)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
		return
	}

	err = g.Initialize()
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
		return
	}
	err = g.WaitForAPIUp()
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
		return
	}
	v = m.Run()
}

func resources(ctx context.Context, h *framework.Host, g *framework.Guest, s *k8s.Setup, helmClient *helmclient.Client) error {
	{
		err := s.EnsureNamespaceCreated(ctx, operatorNamespace)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	//err := h.InstallStableOperator("cert-operator", "certconfig", e2etemplates.CertOperatorChartValues)
	//if err != nil {
	//	return microerror.Mask(err)
	//}
	//err = h.InstallStableOperator("node-operator", "drainerconfig", e2etemplates.NodeOperatorChartValues)
	//if err != nil {
	//	return microerror.Mask(err)
	//}
	//err = h.InstallStableOperator("aws-operator", "awsconfig", e2etemplates.AWSOperatorChartValues)
	//if err != nil {
	//	return microerror.Mask(err)
	//}

	// NOTE that the release package has to be configured to make this work. Right
	// now it is unclear in which direction the e2e tests go but this here should
	// help later. If not, it will just get removed completely.
	//
	//	{
	//		c := chartvalues.E2ESetupVaultConfig{
	//			Vault: chartvalues.E2ESetupVaultConfigVault{
	//				Token: env.VaultToken(),
	//			},
	//		}
	//
	//		values, err := chartvalues.NewE2ESetupVault(c)
	//		if err != nil {
	//			return microerror.Mask(err)
	//		}
	//
	//		err = config.Release.Install(ctx, key.VaultReleaseName(), release.NewStableVersion(), values, config.Release.Condition().PodExists(ctx, "default", "app=vault"))
	//		if err != nil {
	//			return microerror.Mask(err)
	//		}
	//	}
	//
	//	{
	//		c := chartvalues.CertOperatorConfig{
	//			CommonDomain:       env.CommonDomain(),
	//			RegistryPullSecret: env.RegistryPullSecret(),
	//			Vault: chartvalues.CertOperatorVault{
	//				Token: env.VaultToken(),
	//			},
	//		}
	//
	//		values, err := chartvalues.NewCertOperator(c)
	//		if err != nil {
	//			return microerror.Mask(err)
	//		}
	//
	//		err = config.Release.InstallOperator(ctx, key.CertOperatorReleaseName(), release.NewStableVersion(), values, corev1alpha1.NewCertConfigCRD())
	//		if err != nil {
	//			return microerror.Mask(err)
	//		}
	//	}

	err := installCredential(h)
	if err != nil {
		return microerror.Mask(err)
	}

	// NOTE the template got removed but this code uses it still. This cannot
	// work. Not fixing it now as the e2e tests are disabled anyway for this
	// project.
	//
	//	err = h.InstallResource("apiextensions-aws-config-e2e", e2etemplates.ApiextensionsAWSConfigE2EChartValues, ":stable")
	//	if err != nil {
	//		return microerror.Mask(err)
	//	}

	//err = h.InstallBranchOperator("cluster-operator", "awsclusterconfig", template.ClusterOperatorChartValues)
	//if err != nil {
	//	return microerror.Mask(err)
	//}
	//
	//err = h.InstallResource("apiextensions-aws-cluster-config-e2e", template.ClusterOperatorResourceChartValues, ":stable")
	//if err != nil {
	//	return microerror.Mask(err)
	//}

	return nil
}

func installCredential(h *framework.Host) error {
	o := func() error {
		k8sClient := h.K8sClient()

		k8sClient.CoreV1().Secrets(credentialNamespace).Delete(credentialName, &metav1.DeleteOptions{})

		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: credentialName,
			},
			Data: map[string][]byte{
				awsOperatorArnKey: []byte(os.Getenv("GUEST_AWS_ARN")),
			},
		}

		_, err := k8sClient.CoreV1().Secrets(credentialNamespace).Create(secret)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
	n := func(err error, delay time.Duration) {
		log.Println("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
