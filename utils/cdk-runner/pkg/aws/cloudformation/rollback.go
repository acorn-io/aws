package cloudformation

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

func Rollback(c *Client, stackName string) error {
	rollbackWaiter := cloudformation.NewStackRollbackCompleteWaiter(c.Client)
	c.Client.RollbackStack(c.Ctx, &cloudformation.RollbackStackInput{
		StackName: aws.String(stackName),
	})
	if err := rollbackWaiter.Wait(c.Ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}, time.Minute*60); err != nil {
		return err
	}

	return nil
}
