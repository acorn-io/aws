package context

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type CdkContext struct {
	Ec2Client *ec2.Client
	AwsMeta   AwsConfig
	Plugins   []PluginProvider
	Context   context.Context
}

func NewContext(account, region string) (*CdkContext, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	cfg.Region = region
	client := ec2.NewFromConfig(cfg)

	return &CdkContext{
		Ec2Client: client,
		Context:   ctx,
		AwsMeta: AwsConfig{
			Account: account,
			Region:  region,
		},
		Plugins: []PluginProvider{
			NewAzPlugin(),
		},
	}, nil
}

func (ctx *CdkContext) AddPlugin(p PluginProvider) {
	ctx.Plugins = append(ctx.Plugins, p)
}
