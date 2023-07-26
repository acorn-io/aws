#!/bin/bash
set -e

source ./scripts/record_event.sh

STACK_NAME="acorn-${ACORN_ACCOUNT}-${ACORN_PROJECT}-${ACORN_NAME//\./-}"

./scripts/stack_log.sh ${STACK_NAME} &

if [ "${ACORN_EVENT}" = "delete" ]; then
	aws cloudformation delete-stack --stack-name "${STACK_NAME}"
	aws cloudformation wait stack-delete-complete --stack-name "${STACK_NAME}"
	record_event 'S3Deleted' "Bucket successfully deleted"
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
cdk synth --path-metadata false --lookups false >cfn.yaml
cat cfn.yaml

# Run CloudFormation
aws cloudformation deploy --template-file cfn.yaml --stack-name "${STACK_NAME}" --capabilities CAPABILITY_IAM --capabilities CAPABILITY_NAMED_IAM --no-fail-on-empty-changeset --no-cli-pager
aws cloudformation describe-stacks --stack-name "${STACK_NAME}" --query 'Stacks[0].Outputs' >outputs.json

if [ "${ACORN_EVENT}" = "create" ]; then
	record_event 'S3Created' "Bucket successfully created"
fi

if [ "${ACORN_EVENT}" = "update" ]; then
	record_event 'S3Updated' "Bucket successfully updated"
fi

# Render Output
URL="$(jq -r '.[] | select(.OutputKey=="BucketURL") | .OutputValue' outputs.json | cut -d'/' -f3)"
ARN="$(jq -r '.[] | select(.OutputKey=="BucketARN") | .OutputValue' outputs.json)"

cat >/run/secrets/output <<EOF
services: "s3-bucket": {
    default: true
    address: "${URL}"
    data: bucketARN: "${ARN}"
}
EOF
