package elasticache

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// GetPrivateSubnetGroup returns a new subnet group for the given elasticache stack
func GetPrivateSubnetGroup(scope constructs.Construct, name *string, vpc awsec2.IVpc) awselasticache.CfnSubnetGroup {
	privateSubnetIDs := make([]*string, 0)

	for _, subnet := range *vpc.PrivateSubnets() {
		privateSubnetIDs = append(privateSubnetIDs, subnet.SubnetId())
	}

	subnetGroup := awselasticache.NewCfnSubnetGroup(scope, name, &awselasticache.CfnSubnetGroupProps{
		Description:          jsii.String("Acorn created Elasticache subnets"),
		CacheSubnetGroupName: name,
		SubnetIds:            &privateSubnetIDs,
	})

	return subnetGroup
}

// GetAllowAllVPCSecurityGroup returns a security group that allows traffic to and from the Elasticache cluster
func GetAllowAllVPCSecurityGroup(scope constructs.Construct, name *string, vpc awsec2.IVpc, port int) awsec2.SecurityGroup {
	sg := awsec2.NewSecurityGroup(scope, name, &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description:      jsii.String("Acorn created Elasticache security group"),
	})

	for _, i := range *vpc.PrivateSubnets() {
		sg.AddIngressRule(awsec2.Peer_Ipv4(i.Ipv4CidrBlock()), awsec2.Port_Tcp(jsii.Number(port)), jsii.String("Allow from private subnets"), jsii.Bool(false))
	}

	for _, i := range *vpc.PublicSubnets() {
		sg.AddIngressRule(awsec2.Peer_Ipv4(i.Ipv4CidrBlock()), awsec2.Port_Tcp(jsii.Number(port)), jsii.String("Allow from public subnets"), jsii.Bool(false))
	}

	return sg
}
