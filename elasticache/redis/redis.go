package main

import (
	"fmt"

	"github.com/acorn-io/aws/elasticache"
	"github.com/acorn-io/aws/elasticache/redis/util"
	"github.com/acorn-io/services/aws/libs/common"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/sirupsen/logrus"
)

type redisStackProps struct {
	awscdk.StackProps
	ClusterName string
	UserTags    map[string]string `json:"tags" yaml:"tags"`
	VpcID       string
}

func newRedisStack(scope constructs.Construct, props *redisStackProps) (awscdk.Stack, error) {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	// create the stack
	stack := awscdk.NewStack(scope, jsii.String("redis-stack"), &sprops)

	// lookup the VPC
	vpc := awsec2.Vpc_FromLookup(stack, jsii.String("VPC"), &awsec2.VpcLookupOptions{
		VpcId: jsii.String(props.VpcID),
	})

	// get the subnet group
	subnetGroup := elasticache.GetPrivateSubnetGroup(stack, jsii.String(props.ClusterName+"-subnet-group"), vpc)

	// get the security group
	sg := elasticache.GetAllowAllVPCSecurityGroup(stack, jsii.String(props.ClusterName+"-security-group"), vpc, 6379)

	vpcSecurityGroupIDs := make([]*string, 0)
	vpcSecurityGroupIDs = append(vpcSecurityGroupIDs, sg.SecurityGroupId())

	// build up the secrets manager
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	sm := secretsmanager.New(sess)

	// generate the token
	token, err := util.GenerateToken(24)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// store the token
	secretOutput, err := sm.CreateSecret(&secretsmanager.CreateSecretInput{
		Name:         jsii.String(props.ClusterName + "-token"),
		SecretString: jsii.String(token),
		Description:  jsii.String("Acorn generated secret for Redis authentication."),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create token secret: %w", err)
	}

	// create the Redis cluster
	// it might seem like creating a replication group is not the same as creating a cluster
	// but actually it creates the cluster and the replication group in one go
	redisRG := awselasticache.NewCfnReplicationGroup(stack, jsii.String(props.ClusterName), &awselasticache.CfnReplicationGroupProps{
		ReplicationGroupId:          jsii.String(props.ClusterName),
		ReplicationGroupDescription: jsii.String("Acorn created Redis replication group"),
		Engine:                      jsii.String("redis"),
		CacheNodeType:               jsii.String("cache.t4g.micro"),
		// this says num clusters but with cluster mode disabled its actually num nodes
		// also, the terminology is confusing. We're creating an elasticache cluster but not a Redis cluster.
		NumCacheClusters:         jsii.Number(3),
		AutomaticFailoverEnabled: jsii.Bool(true),
		TransitEncryptionEnabled: jsii.Bool(true),
		CacheSubnetGroupName:     subnetGroup.CacheSubnetGroupName(),
		SecurityGroupIds:         &vpcSecurityGroupIDs,
		AuthToken:                jsii.String(token),
		Port:                     jsii.Number(6379),
	})

	// output the cluster details
	awscdk.NewCfnOutput(stack, jsii.String("clustername"), &awscdk.CfnOutputProps{
		Value: jsii.String(props.ClusterName),
	})
	awscdk.NewCfnOutput(stack, jsii.String("address"), &awscdk.CfnOutputProps{
		Value: redisRG.AttrPrimaryEndPointAddress(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("port"), &awscdk.CfnOutputProps{
		Value: redisRG.AttrPrimaryEndPointPort(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("tokenarn"), &awscdk.CfnOutputProps{
		Value: secretOutput.ARN,
	})

	return stack, nil
}

func main() {
	defer jsii.Close()

	app := common.NewAcornTaggedApp(nil)

	stackProps := &redisStackProps{
		StackProps: *common.NewAWSCDKStackProps(),
	}
	stackProps.VpcID = common.GetVpcID()

	err := common.NewConfig(stackProps)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create config")
	}

	id, err := util.GenerateID(8)
	if err != nil {
		logrus.WithError(err).Fatal("failed to generate cluster ID")
	}

	stackProps.ClusterName = stackProps.ClusterName + "-" + id

	common.AppendScopedTags(app, stackProps.UserTags)
	_, err = newRedisStack(app, stackProps)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create Redis stack")
	}

	app.Synth(nil)
}
