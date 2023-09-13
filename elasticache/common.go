package elasticache

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// ResourceID returns an ID that can be used to uniquely identify resources it appends the given suffix to the ID
func ResourceID(suffix string) string {
	return os.Getenv("ACORN_PROJECT") + os.Getenv("ACORN_NAME") + suffix
}

// GetPrivateSubnetGroup returns a new subnet group for the given elasticache stack
func GetPrivateSubnetGroup(scope constructs.Construct, name *string, vpc awsec2.IVpc) awselasticache.CfnSubnetGroup {
	privateSubnetIDs := make([]*string, 0)

	for _, subnet := range *vpc.PrivateSubnets() {
		privateSubnetIDs = append(privateSubnetIDs, subnet.SubnetId())
	}

	subnetGroup := awselasticache.NewCfnSubnetGroup(scope, name, &awselasticache.CfnSubnetGroupProps{
		CacheSubnetGroupName: jsii.String(ResourceID("Sg")),
		Description:          jsii.String("Acorn created Elasticache subnet group."),
		SubnetIds:            &privateSubnetIDs,
	})

	return subnetGroup
}
