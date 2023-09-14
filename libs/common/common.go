package common

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const (
	configFile = "/app/config.json"
)

var (
	acornTags = map[string]string{
		"acorn.io/managed":      "true",
		"acorn.io/project-name": os.Getenv("ACORN_PROJECT"),
		"acorn.io/acorn-name":   os.Getenv("ACORN_NAME"),
		"acorn.io/account-id":   os.Getenv("ACORN_ACCOUNT"),
	}
)

func ConfigBytes() ([]byte, error) {
	return os.ReadFile(configFile)
}

func NewConfig(props any) error {
	conf, err := ConfigBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(conf, props)
}

func NewAcornTaggedApp(props *awscdk.AppProps) awscdk.App {
	app := awscdk.NewApp(props)
	AppendScopedTags(app, acornTags)
	return app
}

func AppendScopedTags(scope constructs.Construct, tags map[string]string) {
	scopedTags := awscdk.Tags_Of(scope)
	for k, v := range tags {
		scopedTags.Add(jsii.String(k), jsii.String(v), &awscdk.TagProps{})
	}
}

func GetVpcID() string {
	return os.Getenv("VPC_ID")
}

func NewAWSCDKStackProps() *awscdk.StackProps {
	return &awscdk.StackProps{
		Synthesizer: awscdk.NewDefaultStackSynthesizer(&awscdk.DefaultStackSynthesizerProps{
			GenerateBootstrapVersionRule: jsii.Bool(false),
		}),
		Env: &awscdk.Environment{
			Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
			Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
		},
	}
}

// GetAllowAllVPCSecurityGroup returns a security group that allows traffic to and from the vpc over the given port
func GetAllowAllVPCSecurityGroup(scope constructs.Construct, name *string, description *string, vpc awsec2.IVpc, port int) awsec2.SecurityGroup {
	sg := awsec2.NewSecurityGroup(scope, name, &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description:      description,
	})

	for _, i := range *vpc.PrivateSubnets() {
		sg.AddIngressRule(awsec2.Peer_Ipv4(i.Ipv4CidrBlock()), awsec2.Port_Tcp(jsii.Number(port)), jsii.String("Allow from private subnets"), jsii.Bool(false))
	}

	for _, i := range *vpc.PublicSubnets() {
		sg.AddIngressRule(awsec2.Peer_Ipv4(i.Ipv4CidrBlock()), awsec2.Port_Tcp(jsii.Number(port)), jsii.String("Allow from public subnets"), jsii.Bool(false))
	}

	return sg
}
