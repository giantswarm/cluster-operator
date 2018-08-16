// +build k8srequired

package setup

import (
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
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/integration/env"
	"github.com/giantswarm/cluster-operator/integration/teardown"
	"github.com/giantswarm/cluster-operator/integration/template"
)

const (
	awsOperatorArnKey   = "aws.awsoperator.arn"
	credentialName      = "credential-default"
	credentialNamespace = "giantswarm"
)

func hostPeerVPC(c *awsclient.Client) (string, error) {
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
		return "", microerror.Mask(err)
	}
	err = c.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return "", microerror.Mask(err)
	}
	describeOutput, err := c.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return "", microerror.Mask(err)
	}

	var vpcPeerID string
	for _, o := range describeOutput.Stacks[0].Outputs {
		if *o.OutputKey == "VPCID" {
			os.Setenv("AWS_VPC_PEER_ID", *o.OutputValue)
			vpcPeerID = *o.OutputValue
			break
		}
	}
	log.Printf("Host Peer VPC stack created")
	return vpcPeerID, nil
}

func WrapTestMain(g *framework.Guest, h *framework.Host, helmClient *helmclient.Client, apprClient *apprclient.Client, m *testing.M) {
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
			name := "aws-operator"
			customResource := "awsconfig"
			logEntry := "deleted the guest cluster main stack"
			h.DeleteGuestCluster(name, customResource, logEntry)

			err := teardown.HostPeerVPC(c)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}

			// only do full teardown when not on CI
			if os.Getenv("CIRCLECI") != "true" {
				err := teardown.Resources(c, h, helmClient)
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

	vpcPeerID, err := hostPeerVPC(c)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
		return
	}

	err = h.Setup()
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
		return
	}

	err = resources(h, g, helmClient, vpcPeerID)
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

func resources(h *framework.Host, g *framework.Guest, helmClient *helmclient.Client, vpcPeerID string) error {
	err := h.InstallStableOperator("cert-operator", "certconfig", e2etemplates.CertOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}
	err = h.InstallStableOperator("node-operator", "drainerconfig", e2etemplates.NodeOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}

	err = installAWSOperator(h)
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

	err = installCredential(h)
	if err != nil {
		return microerror.Mask(err)
	}

	err = installAWSConfig(h, vpcPeerID)
	if err != nil {
		return microerror.Mask(err)
	}

	err = h.InstallBranchOperator("cluster-operator", "awsclusterconfig", template.ClusterOperatorChartValues)
	if err != nil {
		return microerror.Mask(err)
	}

	err = h.InstallResource("apiextensions-aws-cluster-config-e2e", template.ClusterOperatorResourceChartValues, ":stable")
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installAWSOperator(h *framework.Host) error {
	var err error

	var values string
	{
		c := chartvalues.AWSOperatorConfig{
			Provider: chartvalues.AWSOperatorConfigProvider{
				AWS: chartvalues.AWSOperatorConfigProviderAWS{
					Encrypter: "kms",
					Region:    env.AWSRegion(),
				},
			},
			Secret: chartvalues.AWSOperatorConfigSecret{
				AWSOperator: chartvalues.AWSOperatorConfigSecretAWSOperator{
					IDRSAPub: env.IDRSAPub(),
					SecretYaml: chartvalues.AWSOperatorConfigSecretAWSOperatorSecretYaml{
						Service: chartvalues.AWSOperatorConfigSecretAWSOperatorSecretYamlService{
							AWS: chartvalues.AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWS{
								AccessKey: chartvalues.AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWSAccessKey{
									ID:     env.GuestAWSAccessKeyID(),
									Secret: env.GuestAWSAccessKeySecret(),
									Token:  env.GuestAWSAccessKeyToken(),
								},
								HostAccessKey: chartvalues.AWSOperatorConfigSecretAWSOperatorSecretYamlServiceAWSAccessKey{
									ID:     env.HostAWSAccessKeyID(),
									Secret: env.HostAWSAccessKeySecret(),
									Token:  env.HostAWSAccessKeyToken(),
								},
							},
						},
					},
				},
			},
			RegistryPullSecret: env.RegistryPullSecret(),
		}

		values, err = chartvalues.NewAWSOperator(c)
		if err != nil {
			return microerror.Mask(err)
		}

	}

	err = h.InstallStableOperator("aws-operator", "awsconfig", values)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func installAWSConfig(h *framework.Host, vpcPeerID string) error {
	var err error

	var values string
	{
		c := chartvalues.APIExtensionsAWSConfigE2EConfig{
			CommonDomain:         env.CommonDomain(),
			ClusterName:          env.ClusterID(),
			SSHPublicKey:         env.IDRSAPub(),
			VersionBundleVersion: env.VersionBundleVersion(),

			AWS: chartvalues.APIExtensionsAWSConfigE2EConfigAWS{
				Region:            env.AWSRegion(),
				APIHostedZone:     env.AWSAPIHostedZoneGuest(),
				IngressHostedZone: env.AWSIngressHostedZoneGuest(),
				RouteTable0:       env.AWSRouteTable0(),
				RouteTable1:       env.AWSRouteTable1(),
				VPCPeerID:         vpcPeerID,
			},
		}

		values, err = chartvalues.NewAPIExtensionsAWSConfigE2E(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = h.InstallResource("apiextensions-aws-config-e2e", values, ":stable")
	if err != nil {
		return microerror.Mask(err)
	}

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
	b := backoff.NewExponential(framework.ShortMaxWait, framework.ShortMaxInterval)
	n := func(err error, delay time.Duration) {
		log.Println("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
