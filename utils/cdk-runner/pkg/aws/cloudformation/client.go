package cloudformation

import (
	"context"

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
	return &Client{
		Ctx:    ctx,
		Client: cloudformation.NewFromConfig(cfg),
	}, nil
}
