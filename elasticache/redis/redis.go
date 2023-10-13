package main

import (
	"fmt"
	"strings"

	"github.com/acorn-io/aws/elasticache"
	"github.com/acorn-io/services/aws/libs/common"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/sirupsen/logrus"
)

type redisStackProps struct {
	awscdk.StackProps
	ClusterName          string            `json:"clusterName" yaml:"clusterName"`
	UserTags             map[string]string `json:"tags" yaml:"tags"`
	NodeType             string            `json:"nodeType" yaml:"nodeType"`
	NumNodes             int               `json:"numNodes" yaml:"numNodes"`
	SkipSnapshotOnDelete bool              `json:"skipSnapshotOnDelete" yaml:"skipSnapshotOnDelete"`
}

// NewRedisStack creates the new Redis stack
func NewRedisStack(scope constructs.Construct, id string, props *redisStackProps) (awscdk.Stack, error) {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	// create the stack
	stack := awscdk.NewStack(scope, jsii.String(id), &sprops)

	// lookup the VPC
	vpc := awsec2.Vpc_FromLookup(stack, jsii.String("VPC"), &awsec2.VpcLookupOptions{
		VpcId: jsii.String(common.GetVpcID()),
	})

	// get the subnet group
	subnetGroup := elasticache.GetPrivateSubnetGroup(stack, elasticache.ResourceID(props.ClusterName, "Sng"), vpc)

	// get the security group
	sg := common.GetAllowAllVPCSecurityGroup(stack, elasticache.ResourceID(props.ClusterName, "Scg"), jsii.String("Acorn generated Elasticache security group"), vpc, 6379)

	vpcSecurityGroupIDs := make([]*string, 0)
	vpcSecurityGroupIDs = append(vpcSecurityGroupIDs, sg.SecurityGroupId())

	// store the token in the AWS secret manager
	token := awssecretsmanager.NewSecret(stack, jsii.String(props.ClusterName+"Token"), &awssecretsmanager.SecretProps{
		Description: jsii.String("Acorn generated token for Redis authentication."),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			ExcludePunctuation: jsii.Bool(true),
			PasswordLength:     jsii.Number(20),
			IncludeSpace:       jsii.Bool(false),
		},
	})

	// create the Redis cluster
	// it might seem like creating a replication group is not the same as creating a cluster
	// but actually it creates the cluster and the replication group in one go
	redisRG := awselasticache.NewCfnReplicationGroup(stack, jsii.String(props.ClusterName), &awselasticache.CfnReplicationGroupProps{
		ReplicationGroupId:          elasticache.ResourceID(props.ClusterName, ""),
		ReplicationGroupDescription: jsii.String("Acorn created Redis replication group"),
		Engine:                      jsii.String("redis"),
		CacheNodeType:               jsii.String(props.NodeType),
		// this says num clusters but with cluster mode disabled its actually num nodes
		// also, the terminology is confusing. We're creating an elasticache cluster but not a Redis cluster.
		NumCacheClusters:         jsii.Number(props.NumNodes),
		AutomaticFailoverEnabled: jsii.Bool(props.NumNodes > 1),
		TransitEncryptionEnabled: jsii.Bool(true),
		CacheSubnetGroupName:     subnetGroup.CacheSubnetGroupName(),
		SecurityGroupIds:         &vpcSecurityGroupIDs,
		AuthToken:                token.SecretValue().ToString(),
		Port:                     jsii.Number(6379),
		SnapshotRetentionLimit:   jsii.Number(1), // how many days to retain snapshots
	})

	// indicate that the subnet group depends on the cluster
	// this prevents deletion errors caused by attempted subnet group deletes while the cluster still exists
	redisRG.AddDependency(subnetGroup)

	if !props.SkipSnapshotOnDelete {
		// indicate that the cluster should be backed up before deletion
		redisRG.ApplyRemovalPolicy(awscdk.RemovalPolicy_SNAPSHOT, &awscdk.RemovalPolicyOptions{
			ApplyToUpdateReplacePolicy: jsii.Bool(true),
		})
	}

	arn := fmt.Sprintf("arn:aws:elasticache:%s:%s:replicationgroup:%s", *stack.Region(), *stack.Account(), *elasticache.ResourceID(props.ClusterName, ""))
	arn = strings.ToLower(arn)

	// output the cluster details
	awscdk.NewCfnOutput(stack, jsii.String("clustername"), &awscdk.CfnOutputProps{
		Value: jsii.String(props.ClusterName),
	})
	awscdk.NewCfnOutput(stack, jsii.String("clusterarn"), &awscdk.CfnOutputProps{
		Value: jsii.String(arn),
	})
	awscdk.NewCfnOutput(stack, jsii.String("address"), &awscdk.CfnOutputProps{
		Value: redisRG.AttrPrimaryEndPointAddress(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("port"), &awscdk.CfnOutputProps{
		Value: redisRG.AttrPrimaryEndPointPort(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("tokenarn"), &awscdk.CfnOutputProps{
		Value: token.SecretArn(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("transitencryption"), &awscdk.CfnOutputProps{
		Value: jsii.String("true"),
	})

	return stack, nil
}

func main() {
	defer jsii.Close()

	app := common.NewAcornTaggedApp(nil)

	stackProps := &redisStackProps{
		StackProps: *common.NewAWSCDKStackProps(),
	}

	err := common.NewConfig(stackProps)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create config")
	}

	common.AppendScopedTags(app, stackProps.UserTags)
	_, err = NewRedisStack(app, "RedisStack", stackProps)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create Redis stack")
	}

	app.Synth(nil)
}
