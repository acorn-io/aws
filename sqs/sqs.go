package main

import (
	"os"
	"strings"

	"github.com/acorn-io/services/aws/libs/common"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/sirupsen/logrus"
)

type MyStackProps struct {
	awscdk.StackProps
	AccessPolicies            []policyStatement `json:"accessPolicies,omitempty"`
	ContentBasedDeduplication bool              `json:"contentBasedDeduplication,omitempty"`
	DataKeyReuse              int               `json:"dataKeyReuse,omitempty"`
	ExternalID                string
	Fifo                      bool              `json:"fifo,omitempty"`
	MaxReceiveCount           int               `json:"maxReceiveCount,omitempty"`
	QueueName                 string            `json:"queueName,omitempty"`
	UserTags                  map[string]string `json:"tags,omitempty"`
	VisibilityTimeout         int               `json:"visibilityTimeout,omitempty"`
}

// Modified struct based on awsiam.PolicyStatementProps
type policyStatement struct {
	Actions       *[]*string                    `field:"optional" json:"actions" yaml:"actions"`
	Conditions    *map[string]interface{}       `field:"optional" json:"conditions" yaml:"conditions"`
	Effect        awsiam.Effect                 `field:"optional" json:"effect" yaml:"effect"`
	NotActions    *[]*string                    `field:"optional" json:"notActions" yaml:"notActions"`
	NotPrincipals *[]principalFromAcornfileJson `field:"optional" json:"notPrincipals" yaml:"notPrincipals"`
	NotResources  *[]*string                    `field:"optional" json:"notResources" yaml:"notResources"`
	Principals    *[]principalFromAcornfileJson `field:"optional" json:"principals" yaml:"principals"`
	Resources     *[]*string                    `field:"optional" json:"resources" yaml:"resources"`
	Sid           *string                       `field:"optional" json:"sid" yaml:"sid"`
}

// Need this info to get to awsiam.IPrincipal
type principalFromAcornfileJson struct {
	PrincipalType string `json:"principalType,omitempty"`
	Identity      string `json:"identity,omitempty"`
}

func (myStp *MyStackProps) addPoliciesToQueue(queue awssqs.Queue) {
	if len(myStp.AccessPolicies) > 0 {
		for _, policy := range myStp.AccessPolicies {
			principals := []awsiam.IPrincipal{}
			notPrincipals := []awsiam.IPrincipal{}
			if policy.Principals != nil {
				for _, principal := range *policy.Principals {
					if principal.PrincipalType == "service" {
						principals = append(principals, awsiam.NewServicePrincipal(jsii.String(principal.Identity), &awsiam.ServicePrincipalOpts{}))
					}
					if principal.PrincipalType == "arn" {
						principals = append(principals, awsiam.NewArnPrincipal(jsii.String(principal.Identity)))
					}
				}
			}

			if policy.NotPrincipals != nil {
				for _, principal := range *policy.NotPrincipals {
					if principal.PrincipalType == "service" {
						notPrincipals = append(notPrincipals, awsiam.NewServicePrincipal(jsii.String(principal.Identity), &awsiam.ServicePrincipalOpts{}))
					}
					if principal.PrincipalType == "arn" {
						principals = append(principals, awsiam.NewArnPrincipal(jsii.String(principal.Identity)))
					}
				}
			}
			effect := awsiam.Effect_ALLOW
			if policy.Effect == "Deny" {
				effect = awsiam.Effect_DENY
			}
			queue.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Actions:       policy.Actions,
				Conditions:    policy.Conditions,
				Effect:        effect,
				NotActions:    policy.NotActions,
				NotResources:  policy.NotResources,
				Resources:     policy.Resources,
				Sid:           policy.Sid,
				Principals:    &principals,
				NotPrincipals: &notPrincipals,
			}))
		}
	}
}

func NewStackProps() (*MyStackProps, error) {
	stackProps := &MyStackProps{
		StackProps: *common.NewAWSCDKStackProps(),
	}

	if err := common.NewConfig(stackProps); err != nil {
		return nil, err
	}

	if stackProps.VisibilityTimeout == 0 {
		stackProps.VisibilityTimeout = 30
	}

	if stackProps.Fifo && !strings.Contains(stackProps.QueueName, ".fifo") {
		logrus.Infof("Adding required .fifo suffix to queue name: %s", stackProps.QueueName)
		stackProps.QueueName = stackProps.QueueName + ".fifo"
	}

	stackProps.ExternalID = os.Getenv("ACORN_EXTERNAL_ID")

	return stackProps, nil
}

func NewSQSStack(scope constructs.Construct, id string, props *MyStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	queueProps := &awssqs.QueueProps{
		VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(props.VisibilityTimeout)),
		DataKeyReuse:      awscdk.Duration_Seconds(jsii.Number(props.DataKeyReuse)),
	}

	// Workaround for CFN SQS Bug (https://github.com/aws-cloudformation/cloudformation-coverage-roadmap/issues/165)
	if props.Fifo {
		queueProps.Fifo = jsii.Bool(props.Fifo)
		queueProps.ContentBasedDeduplication = jsii.Bool(props.ContentBasedDeduplication)
	}

	if props.QueueName != "" {
		queueProps.QueueName = jsii.String(props.QueueName)
	}

	if props.MaxReceiveCount != 0 {
		dlq := awssqs.NewQueue(stack, jsii.String("sqsQueueDlq"), &awssqs.QueueProps{
			Fifo: jsii.Bool(props.Fifo),
		})
		queueProps.DeadLetterQueue = &awssqs.DeadLetterQueue{
			MaxReceiveCount: jsii.Number(props.MaxReceiveCount),
			Queue:           dlq,
		}
	}

	queue := awssqs.NewQueue(stack, jsii.String("sqsQueue"), queueProps)
	props.addPoliciesToQueue(queue)

	awscdk.NewCfnOutput(stack, jsii.String("QueueURL"), &awscdk.CfnOutputProps{
		Value: queue.QueueUrl(),
	})

	awscdk.NewCfnOutput(stack, jsii.String("QueueARN"), &awscdk.CfnOutputProps{
		Value: queue.QueueArn(),
	})

	awscdk.NewCfnOutput(stack, jsii.String("QueueName"), &awscdk.CfnOutputProps{
		Value: queue.QueueName(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := common.NewAcornTaggedApp(nil)
	stackProps, err := NewStackProps()
	if err != nil {
		logrus.Fatal(err)
	}

	common.AppendScopedTags(app, stackProps.UserTags)
	NewSQSStack(app, "sqsStack", stackProps)

	app.Synth(nil)
}
