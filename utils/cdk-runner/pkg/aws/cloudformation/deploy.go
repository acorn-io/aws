package cloudformation

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/sirupsen/logrus"
)

const (
	StatusFailed                   = "FAILED"
	ReasonNoChanges                = "The submitted information didn't contain changes. Submit different information to create a change set."
	ReasonNoUpdates                = "No updates are to be performed."
	CdkRunnerDeletionProtectionTag = "acorn.io/cdk-runner/deletion-protection"
	DeletionProtectionEnvKey       = "CDK_RUNNER_DELETE_PROTECTION"
)

var (
	acornTags = map[string]string{
		"acorn.io/managed":      "true",
		"acorn.io/project-name": os.Getenv("ACORN_PROJECT"),
		"acorn.io/acorn-name":   os.Getenv("ACORN_NAME"),
		"acorn.io/account-id":   os.Getenv("ACORN_ACCOUNT"),
	}
)

func DeployStack(c *Client, stackName, template string) error {
	if template == "" {
		return fmt.Errorf("template is empty")
	}

	logrus.Infof("Deploying: %s", stackName)

	stack, err := GetStack(c, stackName)
	if err != nil && stack.Exists {
		return err
	}

	// Before doing anything see if we are in a failed state that can be recovered.
	if err := stack.AutoRecover(c); err != nil {
		return err
	}

	changeSetOutput, err := createAndWaitForChangeset(c, stack, template)
	if err != nil || changeSetOutput == nil {
		return err
	}

	if err := outputChangesInChangeSet(c, *changeSetOutput.Id, stack); err != nil {
		return err
	}

	stack.Refresh(c)
	if err := executeChangeSetAndWait(c, *changeSetOutput.Id, stack); err != nil {
		return err
	}
	logrus.Info("Stack Created/Updated")
	return nil
}

func createAndWaitForChangeset(c *Client, stack *CfnStack, template string) (*cloudformation.CreateChangeSetOutput, error) {
	createChangeSetWaiter := cloudformation.NewChangeSetCreateCompleteWaiter(c.Client)

	changeSetType := types.ChangeSetTypeCreate
	if stack.Exists {
		changeSetType = types.ChangeSetTypeUpdate
	}

	tags := getTags()

	logrus.Infof("Creating changeset for: %s", stack.StackName)
	changeSetOutput, err := c.Client.CreateChangeSet(c.Ctx, &cloudformation.CreateChangeSetInput{
		ChangeSetName: aws.String(fmt.Sprintf("%s-%d", stack.StackName, time.Now().Unix())),
		StackName:     aws.String(stack.StackName),
		TemplateBody:  aws.String(template),
		Capabilities: []types.Capability{
			types.CapabilityCapabilityIam,
			types.CapabilityCapabilityNamedIam,
		},
		ChangeSetType: changeSetType,
		Tags:          tags,
	})
	if err != nil {
		return nil, err
	}

	logrus.Infof("Waiting for changeset create/update to complete.")
	if err := createChangeSetWaiter.Wait(c.Ctx, &cloudformation.DescribeChangeSetInput{
		ChangeSetName: changeSetOutput.Id,
		StackName:     aws.String(stack.StackName),
	}, time.Minute*1); err != nil {
		output, describeErr := c.Client.DescribeChangeSet(c.Ctx, &cloudformation.DescribeChangeSetInput{
			ChangeSetName: changeSetOutput.Id,
			StackName:     aws.String(stack.StackName),
		})
		if describeErr != nil {
			return nil, describeErr
		}
		// If a change set is empty, we are not going to fail.
		if output != nil {
			if output.Status == StatusFailed && (strings.Contains(aws.ToString(output.StatusReason), ReasonNoChanges) || strings.Contains(aws.ToString(output.StatusReason), ReasonNoUpdates)) {
				return nil, nil
			}
		}
		return changeSetOutput, err
	}

	return changeSetOutput, nil
}

func outputChangesInChangeSet(c *Client, changeSetId string, stack *CfnStack) error {
	describeChangeSetOutput, err := c.Client.DescribeChangeSet(c.Ctx, &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeSetId),
		StackName:     aws.String(stack.StackName),
	})
	if err != nil {
		return err
	}

	for _, change := range describeChangeSetOutput.Changes {
		logrus.Infof("  %s: %s", change.ResourceChange.Action, *change.ResourceChange.LogicalResourceId)
	}
	return nil
}

func executeChangeSetAndWait(c *Client, changeSetId string, stack *CfnStack) error {
	updateWaiter := cloudformation.NewStackUpdateCompleteWaiter(c.Client)
	createWaiter := cloudformation.NewStackCreateCompleteWaiter(c.Client)

	logrus.Info("Executing changeset")
	if _, err := c.Client.ExecuteChangeSet(c.Ctx, &cloudformation.ExecuteChangeSetInput{
		ChangeSetName: aws.String(changeSetId),
		StackName:     aws.String(stack.StackName),
	}); err != nil {
		return err
	}

	logrus.Info("Waiting for changeset to finish executing")
	if stack.Exists {
		if err := updateWaiter.Wait(c.Ctx, &cloudformation.DescribeStacksInput{
			StackName: aws.String(stack.StackName),
		}, time.Minute*60); err != nil {
			return err
		}
		return nil
	}
	return createWaiter.Wait(c.Ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stack.StackName),
	}, time.Minute*60)
}

func getTags() []types.Tag {
	tags := []types.Tag{}
	for k, v := range acornTags {
		tags = append(tags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	deleteProtection := os.Getenv(DeletionProtectionEnvKey)
	if deleteProtection != "" {
		tags = append(tags, types.Tag{
			Key:   aws.String(CdkRunnerDeletionProtectionTag),
			Value: aws.String(deleteProtection),
		})
	}
	return tags
}
