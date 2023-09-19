package main

import (
	"github.com/acorn-io/aws/elasticache"
	"github.com/acorn-io/services/aws/libs/common"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/sirupsen/logrus"
)

type memcachedStackProps struct {
	awscdk.StackProps
	ClusterName       string            `json:"clusterName" yaml:"clusterName"`
	UserTags          map[string]string `json:"tags" yaml:"tags"`
	NodeType          string            `json:"nodeType" yaml:"nodeType"`
	NumNodes          int               `json:"numNodes" yaml:"numNodes"`
	TransitEncryption bool              `json:"transitEncryption" yaml:"transitEncryption"`
}

// NewMemcachedStack creates the new Memcached stack
func NewMemcachedStack(scope constructs.Construct, id string, props *memcachedStackProps) (awscdk.Stack, error) {
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
	sg := common.GetAllowAllVPCSecurityGroup(stack, elasticache.ResourceID(props.ClusterName, "Scg"), jsii.String("Acorn generated Elasticache security group"), vpc, 11211)

	vpcSecurityGroupIDs := make([]*string, 0)
	vpcSecurityGroupIDs = append(vpcSecurityGroupIDs, sg.SecurityGroupId())

	// create the Memcached cluster
	memcachedCluster := awselasticache.NewCfnCacheCluster(stack, jsii.String(props.ClusterName), &awselasticache.CfnCacheClusterProps{
		ClusterName:              elasticache.ResourceID(props.ClusterName, ""),
		Engine:                   jsii.String("memcached"),
		CacheNodeType:            jsii.String(props.NodeType),
		NumCacheNodes:            jsii.Number(props.NumNodes),
		CacheSubnetGroupName:     subnetGroup.CacheSubnetGroupName(),
		VpcSecurityGroupIds:      &vpcSecurityGroupIDs,
		Port:                     jsii.Number(11211),
		TransitEncryptionEnabled: jsii.Bool(props.TransitEncryption),
	})

	// indicate that the subnet group depends on the cluster
	// this prevents deletion errors caused by attempted subnet group deletes while the cluster still exists
	memcachedCluster.AddDependency(subnetGroup)

	// output the cluster details
	awscdk.NewCfnOutput(stack, jsii.String("clustername"), &awscdk.CfnOutputProps{
		Value: jsii.String(props.ClusterName),
	})
	awscdk.NewCfnOutput(stack, jsii.String("address"), &awscdk.CfnOutputProps{
		Value: memcachedCluster.AttrConfigurationEndpointAddress(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("port"), &awscdk.CfnOutputProps{
		Value: memcachedCluster.AttrConfigurationEndpointPort(),
	})

	return stack, nil
}

func main() {
	defer jsii.Close()

	app := common.NewAcornTaggedApp(nil)

	stackProps := &memcachedStackProps{
		StackProps: *common.NewAWSCDKStackProps(),
	}

	err := common.NewConfig(stackProps)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create config")
	}

	common.AppendScopedTags(app, stackProps.UserTags)
	_, err = NewMemcachedStack(app, "MemcachedStack", stackProps)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create Memcached stack")
	}

	app.Synth(nil)
}
