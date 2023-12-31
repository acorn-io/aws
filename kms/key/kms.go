package main

import (
	"os"
	"strings"

	"github.com/acorn-io/aws/kms/key/props"
	"github.com/acorn-io/services/aws/libs/common"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/sirupsen/logrus"
)

func NewKMSKeyStack(scope constructs.Construct, id string, props *props.KMSKeyStackProps) (awscdk.Stack, error) {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	keySpec, keyUsage, err := props.GetKeySpecAndUsage()
	if err != nil {
		return nil, err
	}

	keyProps := &awskms.KeyProps{
		Enabled:           jsii.Bool(props.Enabled),
		EnableKeyRotation: jsii.Bool(props.EnableKeyRotation),
		KeySpec:           keySpec,
		KeyUsage:          keyUsage,
		PendingWindow:     awscdk.Duration_Days(jsii.Number(props.PendingWindowDays)),

		// Hardcode this to `DESTROY` in order to prevent the user from leaving behind a KMS key that they can't delete.
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	}

	// Set optional properties
	if len(props.KeyAlias) > 0 {
		keyProps.Alias = jsii.String(props.KeyAlias)
	} else {
		keyProps.Alias = jsii.String(strings.ReplaceAll(os.Getenv("DEFAULT_KEY_ALIAS"), ".", "-"))
	}
	if len(props.Description) > 0 {
		keyProps.Description = jsii.String(props.Description)
	}
	if len(props.AdminArn) > 0 {
		keyProps.Admins = &[]awsiam.IPrincipal{awsiam.NewArnPrincipal(jsii.String(props.AdminArn))}
	}
	if len(props.KeyPolicy) > 0 {
		keyProps.Policy = awsiam.PolicyDocument_FromJson(props.KeyPolicy)
	}

	kmsKey := awskms.NewKey(stack, jsii.String(props.KeyName), keyProps)

	awscdk.NewCfnOutput(stack, jsii.String("KMSKeyArn"), &awscdk.CfnOutputProps{
		Value: kmsKey.KeyArn(),
	})

	return stack, nil
}

func main() {
	defer jsii.Close()

	app := common.NewAcornTaggedApp(nil)
	stackProps := &props.KMSKeyStackProps{
		StackProps: *common.NewAWSCDKStackProps(),
	}

	if err := common.NewConfig(stackProps); err != nil {
		logrus.Fatal(err)
	}
	stackProps.SetDefaults()
	if err := stackProps.ValidateProps(); err != nil {
		logrus.Fatalf("invalid stack properties: %s", err)
	}

	common.AppendScopedTags(app, stackProps.Tags)

	if _, err := NewKMSKeyStack(app, "kmsKeyStack", stackProps); err != nil {
		logrus.Fatal(err)
	}

	app.Synth(nil)
}
