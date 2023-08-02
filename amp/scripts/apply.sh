#!/bin/bash

set -e

STACK_NAME="${ACORN_EXTERNAL_ID}"

./scripts/stack_log.sh ${STACK_NAME} &
. ./scripts/record_events.sh

if [ "${ACORN_EVENT}" = "delete" ]; then
    aws cloudformation delete-stack --stack-name "${STACK_NAME}"
    aws cloudformation wait stack-delete-complete --stack-name "${STACK_NAME}"
    record_events "${ACORN_EVENT}d" "Successfully deleted stack ${STACK_NAME}"
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

record_events "${ACORN_EVENT}d" "Successfully applied stack ${STACK_NAME}"

# Render Output
url=$(jq -r '.[] | select(.OutputKey=="AMPEndpointURL")|.OutputValue' outputs.json)
arn=$(jq -r '.[]| select(.OutputKey=="AMPWorkspaceArn")|.OutputValue' outputs.json)
proto="${url%%://*}"
no_proto="${url#*://}"
address="${no_proto%%/*}"
uri="${no_proto#*$address}"

cat > /run/secrets/output<<EOF
services: amp: {
    default: true
    address: "${address}"
    data: {
        arn: "${arn}"
        url: "${url}"
        proto: "${proto}"
        uri: "${uri}"
    }
}
EOF