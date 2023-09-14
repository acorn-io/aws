
# CDK Runner

This project is designed to run CDK in an Acorn environment.

The Runner will:

1. Prepare a cdk.context.json file for the CDK CLI to use and place it in the repo.
1. Run the CDK CLI and output a `cfn.yaml` file.
1. Depending on the Acorn event (create, update, delete) it will either create or delete the stack.
1. If the event is create or update, it will create a change set and call /app/hooks/pre-change-set-apply if it exists.
1. It will write the outputs of the stack to a file called `outputs.json` in the root of the project.
1. It will execute `./scripts/service.sh` if it exists to render the Acorn services.

## Usage

This binary should be copied into the root of the CDK project inside the container you plan to run. It does require that the CDK CLI is installed in the container.
It also assumes that cdk.json will define the `app` key.

The following environment variables are required:
ACORN_EVENT (create, update, delete)
ACORN_EXTERNAL_ID - Will be used to name the CloudFormation stack

See [Needed Environment Variables](#needed-environment-variables) for more details.

## Example

complete examples can be found in the AWS service Acorns.

Here is an abbreviated example of copying the binary into a container with a CDK project.

```Dockerfile
FROM ghcr.io/acorn-io/aws/utils/cdk-runner:v0.1.0 as cdk-runner
FROM cgr.dev/chainguard/wolfi-base
RUN apk add -U --no-cache nodejs bash busybox jq && \
    apk del --no-cache wolfi-base apk-tools
RUN npm install -g aws-cdk
WORKDIR /app
COPY . .
COPY ./cdk.json ./
COPY ./scripts ./scripts
COPY --from=cdk-runner /src/cdk-runner/cdk-runner .
CMD [ "/app/cdk-runner" ]
```

Here is an abbreviated example Acornfile that will run the CDK project

```cue
...
service: "my-service": {
    generated: job: "apply"
}

jobs: apply: {
 build: context: "."
 files: "/app/config.json": std.toJSON(args)
 env: {
  CDK_DEFAULT_ACCOUNT: "@{secrets.aws-context.account-id}"
  CDK_DEFAULT_REGION:  "@{secrets.aws-context.aws-region}"
  VPC_ID:              "@{secrets.aws-context.vpc-id}"
  ACORN_ACCOUNT:       "@{acorn.account}"
  ACORN_NAME:          "@{acorn.name}"
  ACORN_PROJECT:       "@{acorn.project}"
  ACORN_EXTERNAL_ID:   "@{acorn.externalID}"
 }
 events: ["create", "update", "delete"]
 permissions: rules: [
  {
   apiGroup: "aws.acorn.io"
   verbs: [
    "cloudformation:DescribeStacks",
    "cloudformation:CreateChangeSet",
    "cloudformation:DescribeChangeSet",
    "cloudformation:DescribeStackEvents",
    "cloudformation:DescribeStackResources",
    "cloudformation:ExecuteChangeSet",
    "cloudformation:PreviewStackUpdate",
    "cloudformation:UpdateStack",
    "cloudformation:GetTemplateSummary",
    "cloudformation:DeleteStack",
    "aps:*",
   ]
   resources: ["*"]
  }, {
   apiGroup: "aws.acorn.io"
   verbs: [
    "ec2:DescribeAvailabilityZones",
    "ec2:DescribeVpcs",
    "ec2:DescribeSubnets",
    "ec2:DescribeRouteTables",
   ]
   resources: ["*"]
  }, {
   apiGroup: "api.acorn.io"
   verbs: [
    "create",
   ]
   resources: ["events"]
  },

 ]
}

secrets: "aws-context": {
 name:     "AWS Context"
 external: "context://aws"
 type:     "opaque"
 data: {
  "account-id": ""
  "vpc-id":     ""
  "aws-region": ""
 }
}
```

## Needed Environment Variables

- ACORN_EVENT - create, update, delete - This is the event being run and is set by Acorn on the job. This is set by Acorn runtime.
- ACORN_EXTERNAL_ID - This is the external ID of the Acorn and is set by Acorn on the job.
- ACORN_ACCOUNT - This is the account ID of the Acorn and is set by Acorn on the job.
- ACORN_NAME - This is the name of the Acorn and is set by Acorn on the job.
- ACORN_PROJECT - This is the project of the Acorn and is set by Acorn on the job.
- AWS_ROLE_ARN - Set by Acorn identity.
- CDK_DEFAULT_ACCOUNT: AWS account id, required.
- CDK_DEFAULT_REGION:  AWS region, required.
- VPC_ID:              VPC ID, required.

- CDK_RUNNER_DELETE_PROTECTION - Optional, a boolean value that will tag the stack with delete protection enabled. This will be checked when deleting the stack and will not attempt to do so unless it is cleared. To clear it, the Acorn must be updated to disable the deletion protection.

## Hooks

There is a pre-change-set-apply hook that can be used to run scripts before the change set is applied. This is useful for some advanced error processing. The hook is called with the following positional arguments:
The script must be executable and be in the following location:

`/app/hooks/pre-change-set-apply`

The hook is called with the following positional arguments:

1. path to the current applied cloudformation stack template file. (YAML file)
1. path to the new cloudformation stack template file. (YAML file)
1. path to change set json file. (JSON file)
