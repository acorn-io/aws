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

var engine = awsrds.DatabaseClusterEngine_AuroraMysql(&awsrds.AuroraMysqlClusterEngineProps{
	Version: awsrds.AuroraMysqlEngineVersion_VER_3_03_0(),
})

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
		rds.GetAllowAllVPCSecurityGroup(stack, jsii.String("SG"), vpc, 3306),
	}

	creds := awsrds.Credentials_FromGeneratedSecret(jsii.String(props.AdminUser), &awsrds.CredentialsBaseOptions{})

	var parameterGroup awsrds.ParameterGroup
	if len(props.Parameters) > 0 {
		parameterGroup = rds.NewParameterGroup(stack, jsii.String("ParameterGroup"), props, engine)
	}

	cluster := awsrds.NewDatabaseCluster(stack, jsii.String("Cluster"), &awsrds.DatabaseClusterProps{
		Engine:                  engine,
		DefaultDatabaseName:     jsii.String(props.DatabaseName),
		DeletionProtection:      jsii.Bool(props.DeletionProtection),
		CopyTagsToSnapshot:      jsii.Bool(true),
		RemovalPolicy:           rds.GetRemovalPolicy(props),
		Credentials:             creds,
		Vpc:                     vpc,
		SecurityGroups:          sgs,
		ServerlessV2MinCapacity: jsii.Number(props.AuroraCapacityUnitsV2Min),
		ServerlessV2MaxCapacity: jsii.Number(props.AuroraCapacityUnitsV2Max),
		Writer: awsrds.ClusterInstance_ServerlessV2(jsii.String("Instance"), &awsrds.ServerlessV2ClusterInstanceProps{
			EnablePerformanceInsights: jsii.Bool(props.EnablePerformanceInsights),
		}),
		SubnetGroup:    subnetGroup,
		ParameterGroup: parameterGroup,
	})

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
