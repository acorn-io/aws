package rds

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const (
	configFile = "/app/config.json"
)

var (
	SizeMap = map[string]awsec2.InstanceSize{
		"small":  awsec2.InstanceSize_SMALL,
		"medium": awsec2.InstanceSize_MEDIUM,
		"large":  awsec2.InstanceSize_LARGE,
	}
)

type RDSStackProps struct {
	awscdk.StackProps
	DatabaseName       string            `json:"dbName"`
	InstanceSize       string            `json:"instanceSize"`
	AdminUser          string            `json:"adminUsername"`
	Tags               map[string]string `json:"tags"`
	DeletionProtection bool              `json:"deletionProtection"`
	Parameters         map[string]string `json:"parameters"`
	VpcID              string
}

func NewParameterGroup(scope constructs.Construct, name *string, props *RDSStackProps, engine awsrds.IClusterEngine) awsrds.ParameterGroup {
	parameterGroup := awsrds.NewParameterGroup(scope, name, &awsrds.ParameterGroupProps{
		Engine:      engine,
		Description: jsii.String("Acorn created RDS Parameter Group"),
		Parameters:  mapStringToMapStringPtr(props.Parameters),
	})

	return parameterGroup
}

func mapStringToMapStringPtr(from map[string]string) *map[string]*string {
	to := &map[string]*string{}
	for k, v := range from {
		(*to)[k] = jsii.String(v)
	}
	return to
}

func GetAllowAllVPCSecurityGroup(scope constructs.Construct, name *string, vpc awsec2.IVpc) awsec2.SecurityGroup {
	sg := awsec2.NewSecurityGroup(scope, name, &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description:      jsii.String("Acorn created RDS security group"),
	})

	for _, i := range *vpc.PrivateSubnets() {
		sg.AddIngressRule(awsec2.Peer_Ipv4(i.Ipv4CidrBlock()), awsec2.Port_Tcp(jsii.Number(3306)), jsii.String("Allow from private subnets"), jsii.Bool(false))
	}
	for _, i := range *vpc.PublicSubnets() {
		sg.AddIngressRule(awsec2.Peer_Ipv4(i.Ipv4CidrBlock()), awsec2.Port_Tcp(jsii.Number(3306)), jsii.String("Allow from public subnets"), jsii.Bool(false))
	}
	return sg
}

func GetPrivateSubnetGroup(scope constructs.Construct, name *string, vpc awsec2.IVpc) awsrds.SubnetGroup {
	subnetGroup := awsrds.NewSubnetGroup(scope, name, &awsrds.SubnetGroupProps{
		Description: jsii.String("Acorn created RDS Subnets"),
		Vpc:         vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
		},
	})

	return subnetGroup
}
