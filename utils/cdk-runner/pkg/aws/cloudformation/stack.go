package cloudformation

import (
	"strings"

	awsCfn "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/awslabs/goformation"
	"github.com/awslabs/goformation/cloudformation"
	"github.com/sirupsen/logrus"
)

type CfnStack struct {
	StackName          string
	Exists             bool
	Current            types.Stack
	DeletionProtection bool
}

func GetStack(c *Client, stackName string) (*CfnStack, error) {
	cStack := &CfnStack{
		StackName: stackName,
		Exists:    false,
	}

	s, err := c.Client.DescribeStacks(c.Ctx, &awsCfn.DescribeStacksInput{
		StackName: &stackName,
	})
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		logrus.Infof("Stack %s does not exist", stackName)
		return cStack, err
	}

	if len(s.Stacks) > 0 {
		cStack.Exists = s.Stacks[0].StackStatus != types.StackStatusReviewInProgress
		cStack.Current = s.Stacks[0]
		cStack.DeletionProtection = deletionProtectionEnabled(s.Stacks[0])
	}

	return cStack, err
}

func (s *CfnStack) Refresh(c *Client) error {
	stack, err := GetStack(c, s.StackName)
	if err != nil {
		return err
	}
	*s = *stack
	return nil
}

func (s *CfnStack) AutoRecover(c *Client) error {
	if s.Current.StackStatus == types.StackStatusDeleteFailed {
		logrus.Info("Cleaning up failed delete before continuing")
		if err := Delete(c, s.StackName); err != nil {
			return err
		}
	}

	if s.Current.StackStatus == types.StackStatusRollbackFailed && s.Current.DeletionTime != nil {
		logrus.Info("Cleaning up failed rollback/create will delete before continuing")
		if err := Delete(c, s.StackName); err != nil {
			return err
		}
	}

	if s.Current.StackStatus == types.StackStatusRollbackFailed {
		logrus.Info("Cleaning up failed rollback/create will rollback before continuing")
		if err := Rollback(c, s.StackName); err != nil {
			return err
		}
	}

	s.Refresh(c)

	return nil
}

func StackOperationInProgress(c *Client, stackName string) (bool, string, error) {
	stack, err := GetStack(c, stackName)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return false, "", nil
	} else if err != nil {
		return false, "", err
	}
	return strings.Contains(string(stack.Current.StackStatus), "IN_PROGRESS"), string(stack.Current.StackStatus), nil
}

func (s *CfnStack) GetCurrentTemplate(c *Client) (*cloudformation.Template, error) {
	current, err := c.Client.GetTemplate(c.Ctx, &awsCfn.GetTemplateInput{
		StackName: &s.StackName,
	})
	if err != nil {
		return nil, err
	}
	return ParseYAMLCFN([]byte(*current.TemplateBody))
}

func ParseYAMLCFN(template []byte) (*cloudformation.Template, error) {
	return goformation.ParseYAML(template)
}

func deletionProtectionEnabled(stack types.Stack) bool {
	for _, t := range stack.Tags {
		if *t.Key == CdkRunnerDeletionProtectionTag && *t.Value == "true" {
			return true
		}
	}
	return false
}
