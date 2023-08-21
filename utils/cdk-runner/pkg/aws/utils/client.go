package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const (
	AwsRoleArnEnvKey = "AWS_ROLE_ARN"
	AwsSessionEnvKey = "ACORN_EXTERNAL_ID"
)

func WaitForClientRole(ctx context.Context) error {
	timeOutCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	c := sts.NewFromConfig(cfg)
	token, err := os.ReadFile(os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE"))
	if err != nil {
		return err
	}

	for {
		select {
		case <-timeOutCtx.Done():
			return fmt.Errorf("AWS CloudFormation client role not ready after %d seconds", 30)
		default:
			if _, err := c.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
				RoleArn:          aws.String(os.Getenv(AwsRoleArnEnvKey)),
				RoleSessionName:  aws.String(os.Getenv(AwsSessionEnvKey)),
				WebIdentityToken: aws.String(string(token)),
			}); err != nil {
				time.Sleep(time.Second * 2)
				continue
			}
			return nil
		}
	}
}
