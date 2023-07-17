package main

import (
	"strings"

	"github.com/acorn-io/aws/rds"
	"github.com/acorn-io/services/aws/libs/common"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/sirupsen/logrus"
)

type ServerlessV2ScalingAspect struct{}

func (sv2a ServerlessV2ScalingAspect) Visit(node constructs.IConstruct) {
	if n, ok := node.(awsrds.CfnDBCluster); ok {
		n.SetServerlessV2ScalingConfiguration(&awsrds.CfnDBCluster_ServerlessV2ScalingConfigurationProperty{
			MinCapacity: jsii.Number(.5),
			MaxCapacity: jsii.Number(2),
		})
	}
}

func NewServerlessV2ScalingAspect() *ServerlessV2ScalingAspect {
	return &ServerlessV2ScalingAspect{}
}

func NewRDSStack(scope constructs.Construct, props *rds.RDSStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, jsii.String("Stack"), &sprops)

	vpc := awsec2.Vpc_FromLookup(stack, jsii.String("VPC"), &awsec2.VpcLookupOptions{
		VpcId: jsii.String(props.VpcID),
	})

	subnetGroup := rds.GetPrivateSubnetGroup(stack, jsii.String("SubnetGroup"), vpc)

	sgs := &[]awsec2.ISecurityGroup{
		rds.GetAllowAllVPCSecurityGroup(stack, jsii.String("SG"), vpc),
	}

	creds := awsrds.Credentials_FromGeneratedSecret(jsii.String(props.AdminUser), &awsrds.CredentialsBaseOptions{})
	cluster := awsrds.NewDatabaseCluster(stack, jsii.String("Cluster"), &awsrds.DatabaseClusterProps{
		Engine: awsrds.DatabaseClusterEngine_AuroraMysql(&awsrds.AuroraMysqlClusterEngineProps{
			Version: awsrds.AuroraMysqlEngineVersion_VER_3_03_0(),
		}),
		DefaultDatabaseName: jsii.String(props.DatabaseName),
		DeletionProtection:  jsii.Bool(props.DeletionProtection),
		CopyTagsToSnapshot:  jsii.Bool(true),
		RemovalPolicy:       awscdk.RemovalPolicy_SNAPSHOT,
		Credentials:         creds,
		InstanceProps: &awsrds.InstanceProps{
			Vpc:            vpc,
			SecurityGroups: sgs,
			InstanceType:   awsec2.NewInstanceType(jsii.String("serverless")),
		},
		Instances:   jsii.Number(1),
		SubnetGroup: subnetGroup,
	})

	awscdk.Aspects_Of(cluster).Add(NewServerlessV2ScalingAspect())

	port := "3306"
	pSlice := strings.SplitN(*cluster.ClusterEndpoint().SocketAddress(), ":", 2)
	if len(pSlice) == 2 {
		port = pSlice[1]
	}

	awscdk.NewCfnOutput(stack, jsii.String("host"), &awscdk.CfnOutputProps{
		Value: cluster.ClusterEndpoint().Hostname(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("port"), &awscdk.CfnOutputProps{
		Value: &port,
	})
	awscdk.NewCfnOutput(stack, jsii.String("adminusername"), &awscdk.CfnOutputProps{
		Value: creds.Username(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("adminpasswordarn"), &awscdk.CfnOutputProps{
		Value: cluster.Secret().SecretArn(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := common.NewAcornTaggedApp(nil)

	stackProps := &rds.RDSStackProps{
		StackProps: *common.NewAWSCDKStackProps(),
	}
	stackProps.VpcID = common.GetVpcID()

	if err := common.NewConfig(stackProps); err != nil {
		logrus.Fatal(err)
	}

	common.AppendScopedTags(app, stackProps.Tags)
	NewRDSStack(app, stackProps)

	app.Synth(nil)
}
