package context

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type VpcNetworkPlugin struct {
	vpcId string
}

type vpcNetwork struct {
	VpcId           string        `json:"vpcId"`
	VpcCidrBlock    string        `json:"vpcCidrBlock"`
	AvailbiltyZones []string      `json:"availabilityZones"`
	SubnetGroups    []subnetGroup `json:"subnetGroups"`
}

type subnetGroup struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Subnets []subnet `json:"subnets"`
}

type subnet struct {
	SubnetId         string `json:"subnetId"`
	Cidr             string `json:"cidr"`
	AvailabilityZone string `json:"availabilityZone"`
	RouteTableId     string `json:"routeTableId"`
}

func NewVpcPlugin(vpcId string) *VpcNetworkPlugin {
	return &VpcNetworkPlugin{
		vpcId: vpcId,
	}
}

func (v *VpcNetworkPlugin) Render(ctx *CdkContext) (map[string]any, error) {
	key := fmt.Sprintf("vpc-provider:account=%s:filter.vpc-id=%s:region=%s:returnAsymmetricSubnets=true", ctx.AwsMeta.Account, v.vpcId, ctx.AwsMeta.Region)

	d, err := getVpcInfo(v.vpcId, ctx.Context, ctx.Ec2Client)
	return map[string]any{key: d}, err
}

func getVpcInfo(vpcId string, ctx context.Context, c *ec2.Client) (*vpcNetwork, error) {
	resp := &vpcNetwork{
		VpcId:           vpcId,
		SubnetGroups:    []subnetGroup{},
		AvailbiltyZones: []string{},
	}

	vpc, err := c.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcId},
	})
	if err != nil {
		return nil, err
	}

	if len(vpc.Vpcs) > 0 {
		resp.VpcCidrBlock = *vpc.Vpcs[0].CidrBlock
	}

	subnets, err := c.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	routeTables, err := c.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{})
	if err != nil {
		return nil, err
	}

	resp.SubnetGroups = processSubnets(subnets.Subnets, routeTables.RouteTables)
	return resp, nil
}

func groupSubnetByTag(subnets []types.Subnet, tagKeys ...string) (result []types.Subnet) {
	for _, tagKey := range tagKeys {
		for _, subnet := range subnets {
			for _, tag := range subnet.Tags {
				if *tag.Key == tagKey && *tag.Value == "1" {
					result = append(result, subnet)
				}
			}
		}
		if len(result) > 0 {
			return
		}
	}
	return
}

func toSubnets(subnets []types.Subnet, rtbls []types.RouteTable) (result []subnet) {
	for _, s := range subnets {
		result = append(result, subnet{
			SubnetId:         *s.SubnetId,
			Cidr:             *s.CidrBlock,
			AvailabilityZone: *s.AvailabilityZone,
			RouteTableId:     routeTableIdForSubnet(*s.SubnetId, rtbls),
		})
	}
	return
}

func processSubnets(subnets []types.Subnet, routeTables []types.RouteTable) []subnetGroup {
	pubSubnets := groupSubnetByTag(subnets, "subnet.acorn.io/public", "kubernetes.io/role/elb")
	privSubnets := groupSubnetByTag(subnets, "subnet.acorn.io/private", "kubernetes.io/role/internal-elb")

	return []subnetGroup{
		{
			Name:    "acorn-public",
			Type:    "Public",
			Subnets: toSubnets(pubSubnets, routeTables),
		},
		{
			Name:    "acorn-private",
			Type:    "Private",
			Subnets: toSubnets(privSubnets, routeTables),
		},
	}
}

func routeTableIdForSubnet(subnetId string, rtbls []types.RouteTable) string {
	for _, rt := range rtbls {
		for _, asc := range rt.Associations {
			if asc.SubnetId != nil && subnetId == *asc.SubnetId {
				return *rt.RouteTableId
			}
		}
	}
	return ""
}
