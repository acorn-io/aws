package context

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AzPlugin struct{}

func NewAzPlugin() *AzPlugin {
	return &AzPlugin{}
}

func (azp *AzPlugin) Render(ctx *CdkContext) (map[string]any, error) {
	key := fmt.Sprintf("availability-zones:account=%s:region=%s", ctx.AwsMeta.Account, ctx.AwsMeta.Region)
	azInfo, err := getAzInfo(ctx.Context, ctx.Ec2Client)
	return map[string]any{key: azInfo}, err
}

func getAzInfo(ctx context.Context, c *ec2.Client) ([]string, error) {
	azs, err := c.DescribeAvailabilityZones(ctx, &ec2.DescribeAvailabilityZonesInput{})
	if err != nil {
		return nil, err
	}

	var zones []string
	for _, i := range azs.AvailabilityZones {
		zones = append(zones, *i.ZoneName)
	}

	return zones, nil
}
