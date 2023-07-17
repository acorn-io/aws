package main

import (
	"os"

	"github.com/acorn-io/services/aws/libs/common"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsaps"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/sirupsen/logrus"
)

type AmpStackProps struct {
	StackProps    awscdk.StackProps
	Tags          map[string]string `json:"tags"`
	WorkspaceName string            `json:"workspaceName"`
}

func (aps *AmpStackProps) setWorkspaceName() {
	if aps.WorkspaceName == "" {
		aps.WorkspaceName = os.Getenv("ACORN_WORKSPACE")
	}
}

func NewAmpStack(scope constructs.Construct, id string, props *AmpStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	ampWorkspace := awsaps.NewCfnWorkspace(stack, jsii.String(props.WorkspaceName), &awsaps.CfnWorkspaceProps{
		Alias: jsii.String(props.WorkspaceName),
	})

	awscdk.NewCfnOutput(stack, jsii.String("AMPEndpointURL"), &awscdk.CfnOutputProps{
		Value: ampWorkspace.AttrPrometheusEndpoint(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("AMPWorkspaceArn"), &awscdk.CfnOutputProps{
		Value: ampWorkspace.AttrArn(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := common.NewAcornTaggedApp(nil)
	stackProps := &AmpStackProps{
		StackProps: *common.NewAWSCDKStackProps(),
	}

	if err := common.NewConfig(stackProps); err != nil {
		logrus.Fatal(err)
	}
	stackProps.setWorkspaceName()

	common.AppendScopedTags(app, stackProps.Tags)

	NewAmpStack(app, "ampStack", stackProps)

	app.Synth(nil)
}
