package main

import (
	"github.com/acorn-io/services/aws/libs/common"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/sirupsen/logrus"
)

type MyStackProps struct {
	awscdk.StackProps
	MakePublic bool              `json:"makePublic" yaml:"makePublic"`
	Versioned  bool              `json:"versioned" yaml:"versioned"`
	BucketName string            `json:"bucketName" yaml:"bucketName"`
	UserTags   map[string]string `json:"tags" yaml:"tags"`
}

func NewMyStack(scope constructs.Construct, id string, props *MyStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, jsii.String(id), &sprops)

	// Create an S3 bucket
	bucket := awss3.NewBucket(stack, jsii.String(props.BucketName), &awss3.BucketProps{
		Versioned:     jsii.Bool(props.Versioned),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	// Output the bucket URL and ARN
	awscdk.NewCfnOutput(stack, jsii.String("BucketURL"), &awscdk.CfnOutputProps{
		Value: bucket.BucketWebsiteUrl(),
	})

	awscdk.NewCfnOutput(stack, jsii.String("BucketARN"), &awscdk.CfnOutputProps{
		Value: bucket.BucketArn(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := common.NewAcornTaggedApp(nil)

	stackProps := &MyStackProps{
		StackProps: *common.NewAWSCDKStackProps(),
	}

	// Read config from file
	if err := common.NewConfig(stackProps); err != nil {
		logrus.Fatal(err)
	}

	common.AppendScopedTags(app, stackProps.UserTags)

	NewMyStack(app, "s3Stack", stackProps)

	app.Synth(nil)
}
