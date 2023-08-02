#!/bin/bash

STACK_NAME="${ACORN_EXTERNAL_ID}"

# Start logging
./scripts/stacklog.sh ${STACK_NAME} &
. ./scripts/record_event.sh

set -e -x

cdk_synth() {
    cat cdk.context.json
    cdk synth --path-metadata false --lookups false > cfn.yaml
    cat cfn.yaml
}

event_success() {
  record_event "Service${ACORN_EVENT^}d" "CFN Stack: ${STACK_NAME} ${ACORN_EVENT}d successfully"
}

apply_and_render() {
  # Run CDK synth
  cdk_synth
  
  # Run CloudFormation
  aws cloudformation deploy --template-file cfn.yaml --stack-name "${STACK_NAME}" --capabilities CAPABILITY_IAM --capabilities CAPABILITY_NAMED_IAM --no-fail-on-empty-changeset --no-cli-pager
  aws cloudformation describe-stacks --stack-name "${STACK_NAME}" --query 'Stacks[0].Outputs' > outputs.json

  # Render Output
  PORT="$(          jq -r '.[] | select(.OutputKey=="port")            |.OutputValue' outputs.json )"
  ADDRESS="$(       jq -r '.[] | select(.OutputKey=="host")            |.OutputValue' outputs.json )"
  ADMIN_USERNAME="$(jq -r '.[] | select(.OutputKey=="adminusername")   |.OutputValue' outputs.json )"
  PASSWORD_ARN="$(  jq -r '.[] | select(.OutputKey=="adminpasswordarn")|.OutputValue' outputs.json )"

  # Turn off echo
  set +x
  ADMIN_PASSWORD="$(aws --output json secretsmanager get-secret-value --secret-id "${PASSWORD_ARN}" --query 'SecretString' | jq -r .|jq -r .password)"
  if [ "${ACORN_EVENT}" != "delete" ]; then
    event_success
  fi

  cat > /run/secrets/output <<EOF
services: rds: {
  default: true
  address: "${ADDRESS}"
  ports: [${PORT}]
  data: dbName: "${DB_NAME}"
}

secrets: "admin": {
	type: "basic"
	data: {
    username: "${ADMIN_USERNAME}"
    password: "${ADMIN_PASSWORD}"
  }
}
EOF

  if [ -z "${DB_USERNAME}" ]; then
    echo 'services: rds: secrets: ["admin"]' >> /run/secrets/output
  else
    echo 'services: rds: secrets: ["admin", "user"]' >> /run/secrets/output
  fi
}

delete_stack() {
    apply_and_render
    if $(grep 'DeletionProtection: true' cfn.yaml > /dev/null 2>&1); then
        echo "DeletionProtection is enabled, update acorn app with '--deletion-protection=false' to delete this stack..."
        exit 1
    fi

    # Run CloudFormation
    aws cloudformation delete-stack --stack-name "${STACK_NAME}"
    aws cloudformation wait stack-delete-complete --stack-name "${STACK_NAME}" --no-cli-pager
    event_success
}

if [ "${ACORN_EVENT}" = "delete" ]; then
    delete_stack
    exit 0
fi

# Delete failing stacks
STATUS=$(aws cloudformation describe-stacks --stack-name "${STACK_NAME}" | jq -r '.Stacks[0].StackStatus')

if [ "$STATUS" = "DELETE_FAILED" ]; then
    delete_stack
fi

# Run CloudFormation
apply_and_render
exit 0