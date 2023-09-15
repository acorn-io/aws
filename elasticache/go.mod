module github.com/acorn-io/aws/elasticache

go 1.21

require (
	github.com/acorn-io/services/aws/libs/common v0.0.0
	github.com/aws/aws-cdk-go/awscdk v1.204.0-devpreview
	github.com/aws/aws-cdk-go/awscdk/v2 v2.87.0
	github.com/aws/constructs-go/constructs/v10 v10.2.69
	github.com/aws/jsii-runtime-go v1.88.0
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/acorn-io/services/aws/libs/common v0.0.0 => ./../libs/common

require (
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/aws/constructs-go/constructs/v3 v3.4.232 // indirect
	github.com/cdklabs/awscdk-asset-awscli-go/awscliv1/v2 v2.2.198 // indirect
	github.com/cdklabs/awscdk-asset-kubectl-go/kubectlv20/v2 v2.1.2 // indirect
	github.com/cdklabs/awscdk-asset-node-proxy-agent-go/nodeproxyagentv5/v2 v2.0.165 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/yuin/goldmark v1.4.13 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/tools v0.12.0 // indirect
)