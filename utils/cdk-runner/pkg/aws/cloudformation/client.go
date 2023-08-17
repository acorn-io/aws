package cloudformation

import (
	"context"

	"github.com/acorn-io/aws/utils/cdk-runner/pkg/aws/utils"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

type Client struct {
	Ctx    context.Context
	Client *cloudformation.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Ctx:    ctx,
		Client: cloudformation.NewFromConfig(cfg),
	}

	if err := utils.WaitForClientRole(ctx); err != nil {
		return nil, err
	}

	return client, nil
}
