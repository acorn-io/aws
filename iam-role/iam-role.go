package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/acorn-io/services/aws/libs/common"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/sirupsen/logrus"
)

type IAMRoleStackProps struct {
	StackProps                awscdk.StackProps
	Tags                      map[string]string `json:"tags"`
	RoleName                  string            `json:"roleName"`
	TrustedArn                string            `json:"trustedArn"`
	MaxSessionDurationMinutes int               `json:"maxSessionDurationMinutes"`
	Path                      string            `json:"path"`
	ExternalIds               string            `json:"externalIds"`
}

func (rsp *IAMRoleStackProps) setDefaults() {
	if rsp.RoleName == "" {
		rsp.RoleName = os.Getenv("ACORN_EXTERNAL_ID")
	}
	if rsp.MaxSessionDurationMinutes == 0 {
		rsp.MaxSessionDurationMinutes = 60
	}
	if rsp.Path == "" {
		rsp.Path = "/"
	}
}

func (rsp *IAMRoleStackProps) validateProps(policy string) error {
	var errs []error
	if rsp.MaxSessionDurationMinutes < 60 {
		errs = append(errs, fmt.Errorf("maxSessionDurationMinutes must be at least 60"))
	}
	if !strings.HasPrefix(rsp.Path, "/") {
		errs = append(errs, fmt.Errorf("path must start with a /"))
	}
	if policy == "" {
		errs = append(errs, fmt.Errorf("policy cannot be empty"))
	}
	if _, err := arn.Parse(rsp.TrustedArn); err != nil {
		errs = append(errs, fmt.Errorf("failed to parse trustedArn: %w", err))
	}
	return errors.Join(errs...)
}

func NewIAMRoleStack(scope constructs.Construct, id string, props *IAMRoleStackProps, policy []byte) (awscdk.Stack, error) {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	var policyMap map[string]interface{}
	if err := json.Unmarshal(policy, &policyMap); err != nil {
		return nil, fmt.Errorf("error unmarshaling policy document: %w", err)
	}

	roleProps := &awsiam.RoleProps{
		AssumedBy:          awsiam.NewArnPrincipal(jsii.String(props.TrustedArn)),
		Description:        jsii.String("Acorn created IAM Role"),
		InlinePolicies:     &map[string]awsiam.PolicyDocument{"inline": awsiam.PolicyDocument_FromJson(policyMap)},
		MaxSessionDuration: awscdk.Duration_Minutes(jsii.Number(props.MaxSessionDurationMinutes)),
		Path:               jsii.String(props.Path),
		RoleName:           jsii.String(props.RoleName),
	}
	if props.ExternalIds != "" {
		externalIds := make([]*string, len(strings.Split(props.ExternalIds, ",")))
		for i, v := range strings.Split(props.ExternalIds, ",") {
			externalIds[i] = jsii.String(v)
		}

		roleProps.ExternalIds = &externalIds
	}

	iamRole := awsiam.NewRole(stack, jsii.String(props.RoleName), roleProps)

	awscdk.NewCfnOutput(stack, jsii.String("IAMRoleArn"), &awscdk.CfnOutputProps{
		Value: iamRole.RoleArn(),
	})

	return stack, nil
}

func main() {
	defer jsii.Close()

	app := common.NewAcornTaggedApp(nil)
	stackProps := &IAMRoleStackProps{
		StackProps: *common.NewAWSCDKStackProps(),
	}

	policy, err := os.ReadFile("/app/policy.json")
	if err != nil {
		logrus.Fatal(err)
	}

	if err := common.NewConfig(stackProps); err != nil {
		logrus.Fatal(err)
	}
	stackProps.setDefaults()
	if err := stackProps.validateProps(string(policy)); err != nil {
		logrus.Fatalf("invalid stack properties: %w", err)
	}

	common.AppendScopedTags(app, stackProps.Tags)

	if _, err := NewIAMRoleStack(app, "iamRoleStack", stackProps, policy); err != nil {
		logrus.Fatal(err)
	}

	app.Synth(nil)
}
