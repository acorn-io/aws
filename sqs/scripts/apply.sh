#!/bin/bash
set -e

STACK_NAME="${ACORN_EXTERNAL_ID}"

./scripts/stacklog.sh ${STACK_NAME} &
. ./scripts/record_event.sh

record_success() {
    record_event "Service${ACORN_EVENT^}d" "CFN Stack: ${STACK_NAME} ${ACORN_EVENT}d successfully"
}

if [ "${ACORN_EVENT}" = "delete" ]; then
    aws cloudformation delete-stack --stack-name "${STACK_NAME}"
    aws cloudformation wait stack-delete-complete --stack-name "${STACK_NAME}"
    record_success
    exit 0
fi

# Delete failing stacks
STATUS=$(aws cloudformation describe-stacks --stack-name "${STACK_NAME}" | jq -r '.Stacks[0].StackStatus')

if [ "$STATUS" = "DELETE_FAILED" ]; then
    aws cloudformation delete-stack --stack-name "${STACK_NAME}"
    aws cloudformation wait stack-delete-complete --stack-name "${STACK_NAME}" --no-cli-pager
fi

# Run CDK synth
cat cdk.context.json
cdk synth --path-metadata false --lookups false > cfn.yaml
cat cfn.yaml

# Run CloudFormation
aws cloudformation deploy --template-file cfn.yaml --stack-name "${STACK_NAME}" --capabilities CAPABILITY_IAM --capabilities CAPABILITY_NAMED_IAM --no-fail-on-empty-changeset --no-cli-pager
aws cloudformation describe-stacks --stack-name "${STACK_NAME}" --query 'Stacks[0].Outputs' > outputs.json

# Render Output
url="$(jq -r '.[] | select(.OutputKey=="QueueURL")   |.OutputValue' outputs.json |cut -d'/' -f3)"
arn="$(jq -r '.[] | select(.OutputKey=="QueueARN")|.OutputValue' outputs.json )"
name="$(jq -r '.[] | select(.OutputKey=="QueueName")|.OutputValue' outputs.json )"

proto="${url%%://*}"
no_proto="${url#*://}"
address="${no_proto%%/*}"
uri="${no_proto#*$address}"
record_success


cat > /run/secrets/output <<EOF
services: "sqs-queue": {
    default: true
    address: "${address}"
    data: {
        arn: "${arn}"
        proto: "${proto}"
        url: "${url}"
        uri: "${uri}"
        name: "${name}"
    }
}
EOF
