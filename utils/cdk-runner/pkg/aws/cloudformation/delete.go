package cloudformation

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/sirupsen/logrus"
)

func Delete(c *Client, stackName string) error {
	deleteStackWaiter := cloudformation.NewStackDeleteCompleteWaiter(c.Client)

	logrus.Infof("Deleting stack %s", stackName)

	c.Client.DeleteStack(c.Ctx, &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})

	return deleteStackWaiter.Wait(c.Ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}, time.Minute*60)
}
