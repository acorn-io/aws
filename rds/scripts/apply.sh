#!/bin/bash
set -e -x

STACK_NAME="acorn-${ACORN_ACCOUNT}-${ACORN_PROJECT}-${ACORN_NAME//\./-}-${DB_NAME//\./-}"

# Start logging
./scripts/stacklog.sh ${STACK_NAME} &

if [ "${ACORN_EVENT}" = "delete" ]; then
    aws cloudformation delete-stack --stack-name "${STACK_NAME}"
    aws cloudformation wait stack-delete-complete --stack-name "${STACK_NAME}" --no-cli-pager
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
PORT="$(          jq -r '.[] | select(.OutputKey=="port")            |.OutputValue' outputs.json )"
ADDRESS="$(       jq -r '.[] | select(.OutputKey=="host")            |.OutputValue' outputs.json )"
ADMIN_USERNAME="$(jq -r '.[] | select(.OutputKey=="adminusername")   |.OutputValue' outputs.json )"
PASSWORD_ARN="$(  jq -r '.[] | select(.OutputKey=="adminpasswordarn")|.OutputValue' outputs.json )"

# Turn off echo
set +x
ADMIN_PASSWORD="$(aws --output json secretsmanager get-secret-value --secret-id "${PASSWORD_ARN}" --query 'SecretString' | jq -r .|jq -r .password)"

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

if [ -z "$MYSQL_USER" ]; then
  echo 'services: rds: secrets: ["admin"]' >> /run/secrets/output
else
  echo 'services: rds: secrets: ["admin", "user"]' >> /run/secrets/output
fi
