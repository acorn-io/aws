package cloudformation

import (
	"fmt"
	"os"
	"time"

	"github.com/acorn-io/aws/utils/cdk-runner/pkg/cdk"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/sirupsen/logrus"
)

func Delete(c *Client, stackName string) error {
	deleteStackWaiter := cloudformation.NewStackDeleteCompleteWaiter(c.Client)

	// Check if stack exists
	stack, err := GetStack(c, stackName)
	if !stack.Exists {
		// Doesn't exist and user is trying to delete, so we're good
		return nil
	} else if err != nil {
		return err
	}

	if stack.DeletionProtection && os.Getenv(DeletionProtectionEnvKey) == "true" {
		return fmt.Errorf("stack %s has deletion protection enabled, please disable before deleting", stackName)
	} else if stack.DeletionProtection && os.Getenv(DeletionProtectionEnvKey) != "true" {
		if err := cdk.GenerateTemplateFile("cfn.yaml"); err != nil {
			return err
		}
		templateBytes, err := os.ReadFile("cfn.yaml")
		if err != nil {
			return err
		}
		if err := DeployStack(c, stackName, string(templateBytes)); err != nil {
			return err
		}
	}

	logrus.Infof("Deleting stack %s", stackName)

	if _, err := c.Client.DeleteStack(c.Ctx, &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	}); err != nil {
		return err
	}

	stack, err := GetStack(c, stackName)
	if err != nil && stack.Exists {
		return err
	} else if !stack.Exists {
		// Return nil, since we wanted it deleted.
		return nil
	}
	go stack.LogEvents(c)

	return deleteStackWaiter.Wait(c.Ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}, time.Minute*60)
}
