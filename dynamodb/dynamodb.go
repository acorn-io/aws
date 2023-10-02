package main

import (
	"github.com/acorn-io/services/aws/libs/common"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/sirupsen/logrus"
)

type DynamoStackProps struct {
	awscdk.StackProps
	TableName            string            `json:"tableName" yaml:"tableName"`
	PartitionKey         string            `json:"partitionKey" yaml:"partitionKey"`
	PartitionKeyType     string            `json:"partitionKeyType" yaml:"partitionKeyType"`
	SortKey              string            `json:"sortKey" yaml:"sortKey"`
	SortKeyType          string            `json:"sortKeyType" yaml:"sortKeyType"`
	UserTags             map[string]string `json:"tags" yaml:"tags"`
	SkipSnapshotOnDelete bool              `json:"skipSnapshotOnDelete" yaml:"skipSnapshotOnDelete"`
}

func mustGetAttributeType(attrType, arg string) awsdynamodb.AttributeType {
	switch attrType {
	case "STRING":
		return awsdynamodb.AttributeType_STRING
	case "BINARY":
		return awsdynamodb.AttributeType_BINARY
	case "NUMBER":
		return awsdynamodb.AttributeType_NUMBER
	}

	logrus.WithField("arg", arg).Fatalf("unmatched attribute type: %s. Valid values are STRING, BINARY, and NUMBER.", attrType)
	return "" // this won't be hit given that the fatal log causes a panic
}

func NewDynamoStack(scope constructs.Construct, id string, props *DynamoStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, jsii.String(id), &sprops)

	tableProps := &awsdynamodb.TablePropsV2{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String(props.PartitionKey),
			Type: mustGetAttributeType(props.PartitionKeyType, "partitionKeyType"),
		},
	}

	if len(props.TableName) > 0 {
		tableProps.TableName = jsii.String(props.TableName)
	}

	if props.SkipSnapshotOnDelete {
		tableProps.RemovalPolicy = awscdk.RemovalPolicy_DESTROY
	} else {
		tableProps.RemovalPolicy = awscdk.RemovalPolicy_SNAPSHOT
	}

	if len(props.SortKey) > 0 && len(props.SortKeyType) > 0 {
		tableProps.SortKey = &awsdynamodb.Attribute{
			Name: jsii.String(props.SortKey),
			Type: mustGetAttributeType(props.SortKeyType, "sortKeyType"),
		}
	}

	table := awsdynamodb.NewTableV2(stack, jsii.String("ddb-id"), tableProps)

	awscdk.NewCfnOutput(stack, jsii.String("TableName"), &awscdk.CfnOutputProps{
		Value: table.TableName(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("TableARN"), &awscdk.CfnOutputProps{
		Value: table.TableArn(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := common.NewAcornTaggedApp(nil)

	stackProps := &DynamoStackProps{
		StackProps: *common.NewAWSCDKStackProps(),
	}

	err := common.NewConfig(stackProps)
	if err != nil {
		logrus.Fatal(err)
	}

	common.AppendScopedTags(app, stackProps.UserTags)
	NewDynamoStack(app, "dynamoDbStack", stackProps)
	app.Synth(nil)
}
