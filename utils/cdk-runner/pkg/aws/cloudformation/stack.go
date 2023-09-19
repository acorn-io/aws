package cloudformation

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/acorn-io/aws/utils/cdk-runner/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfn "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
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

	if s != nil && len(s.Stacks) > 0 {
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

	if s.Current.StackStatus == types.StackStatusRollbackComplete && s.Current.DeletionTime != nil {
		logrus.Info("Cleaning up rolled back and deleted stack, will delete before continuing")
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

	// We should be able to recover from this one.
	if stack.Current.StackStatus == types.StackStatusReviewInProgress {
		return false, string(stack.Current.StackStatus), nil
	}

	return strings.Contains(string(stack.Current.StackStatus), "IN_PROGRESS"), string(stack.Current.StackStatus), nil
}

func (s *CfnStack) LogEvents(c *Client) {
	var startTime time.Time
	termMessage := strings.Builder{}
	for {
		events, err := c.Client.DescribeStackEvents(c.Ctx, &awsCfn.DescribeStackEventsInput{
			StackName: &s.StackName,
		})
		if err != nil && strings.Contains(err.Error(), "does not exist") {
			time.Sleep(5 * time.Second)
			continue
		} else if err != nil {
			logrus.Error(err)
			continue
		}

		sort.SliceStable(events.StackEvents, func(i, j int) bool {
			timeI := *events.StackEvents[i].Timestamp
			timeJ := *events.StackEvents[j].Timestamp
			return timeI.Before(timeJ)
		})

		for _, event := range events.StackEvents {
			if event.Timestamp.After(startTime) {
				logrus.Infof("%s %s %s %s", event.Timestamp.Format(time.RFC3339), aws.ToString(event.LogicalResourceId), event.ResourceStatus, aws.ToString(event.ResourceStatusReason))
				if event.ResourceStatus == types.ResourceStatusCreateFailed || event.ResourceStatus == types.ResourceStatusUpdateFailed || event.ResourceStatus == types.ResourceStatusDeleteFailed {
					termMessage.WriteString(fmt.Sprintf("%s %s %s\n", aws.ToString(event.LogicalResourceId), event.ResourceStatus, aws.ToString(event.ResourceStatusReason)))
				}
				startTime = *event.Timestamp
			}
		}
		utils.WriteToTermLogAndError([]byte(termMessage.String()), nil)
		time.Sleep(5 * time.Second)
	}
}

func (s *CfnStack) GetCurrentTemplate(c *Client) ([]byte, error) {
	current, err := c.Client.GetTemplate(c.Ctx, &awsCfn.GetTemplateInput{
		StackName: &s.StackName,
	})
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return []byte{}, nil
	} else if err != nil {
		return nil, err
	}

	return []byte(*current.TemplateBody), nil
}

func deletionProtectionEnabled(stack types.Stack) bool {
	for _, t := range stack.Tags {
		if *t.Key == CdkRunnerDeletionProtectionTag && *t.Value == "true" {
			return true
		}
	}
	return false
}
