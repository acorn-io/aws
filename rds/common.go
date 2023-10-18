package rds

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

var (
	InstanceSizeMap = map[string]awsec2.InstanceSize{
		"small":   awsec2.InstanceSize_SMALL,
		"medium":  awsec2.InstanceSize_MEDIUM,
		"large":   awsec2.InstanceSize_LARGE,
		"xlarge":  awsec2.InstanceSize_XLARGE,
		"2xlarge": awsec2.InstanceSize_XLARGE2,
	}
	ComputeClassMap = map[string]awsec2.InstanceClass{
		"burstable":               awsec2.InstanceClass_BURSTABLE3,
		"burstableGraviton":       awsec2.InstanceClass_BURSTABLE4_GRAVITON,
		"standard":                awsec2.InstanceClass_M5,
		"standardGraviton":        awsec2.InstanceClass_M7G,
		"memoryOptimized":         awsec2.InstanceClass_R5,
		"memoryOptimizedGraviton": awsec2.InstanceClass_R7G,
	}
)

type RDSStackProps struct {
	awscdk.StackProps
	AdminUser                 string            `json:"adminUsername"`
	DatabaseName              string            `json:"dbName"`
	DeletionProtection        bool              `json:"deletionProtection"`
	EnablePerformanceInsights bool              `json:"enablePerformanceInsights"`
	InstanceClass             string            `json:"instanceClass"`
	InstanceSize              string            `json:"instanceSize"`
	Parameters                map[string]string `json:"parameters"`
	RestoreSnapshotArn        string            `json:"restoreFromSnapshotArn"`
	SkipSnapShotOnDelete      bool              `json:"skipSnapshotOnDelete"`
	Tags                      map[string]string `json:"tags"`
	VpcID                     string
	// Scaling units for serverless v1
	AuroraCapacityUnitsMin   int `json:"auroraCapacityUnitsMin"`
	AuroraCapacityUnitsMax   int `json:"auroraCapacityUnitsMax"`
	AutoPauseDurationMinutes int `json:"autoPauseDurationMinutes"`
	// Scaling Units for serverless v2
	AuroraCapacityUnitsV2Min float64 `json:"auroraCapacityUnitsV2Min"`
	AuroraCapacityUnitsV2Max float64 `json:"auroraCapacityUnitsV2Max"`
}

type SnapshotAspect struct {
	SnapshotIdentifier string
}

func (sa *SnapshotAspect) Visit(node constructs.IConstruct) {
	if n, ok := node.(awsrds.CfnDBCluster); ok {
		n.AddPropertyOverride(jsii.String("SnapshotIdentifier"), jsii.String(sa.SnapshotIdentifier))
	}
}

func NewSnapshotAspect(snapshotIdentifier string) *SnapshotAspect {
	return &SnapshotAspect{
		SnapshotIdentifier: snapshotIdentifier,
	}
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

func ValidInstanceParameters(instanceClass string, instanceSize string) bool {
	if _, ok := InstanceSizeMap[instanceSize]; !ok {
		return false
	}
	if _, ok := ComputeClassMap[instanceClass]; !ok {
		return false
	}
	return true
}

func GetRemovalPolicy(props *RDSStackProps) awscdk.RemovalPolicy {
	if props.SkipSnapShotOnDelete {
		return awscdk.RemovalPolicy_DESTROY
	}
	return awscdk.RemovalPolicy_SNAPSHOT
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
