#!/bin/bash
STACK_NAME="${ACORN_EXTERNAL_ID}"

echo "Running ${ACORN_EVENT} job event on ${STACK_NAME}"

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

get_current_applied_template() {
  aws cloudformation get-template --stack-name "${STACK_NAME}" --query 'TemplateBody' --output json | jq -r . > current_applied_template.yaml
  python -c 'import yaml, json, sys; print(json.dumps(yaml.safe_load(sys.stdin)))' < current_applied_template.yaml | jq -r . > current-cfn.json
  python -c 'import yaml, json, sys; print(json.dumps(yaml.safe_load(sys.stdin)))' < cfn.yaml | jq -r . > cfn.json
}

## This is a workaround because retention policy updates are treated as no-op by CFN. Look forward to deleting this.
update_stack_for_deletion_policy_only_updates() {
  if [ "$(aws cloudformation list-change-sets --stack-name ${STACK_NAME} | jq '.Summaries[] | "\(.CreationTime):\(.Status):\(.StatusReason)"'|sort -r | head -n 1 |grep "didn't contain changes" |wc -l)" != "1" ]; then
    echo "Was not an empty change set... exiting"
    exit 0
  fi
  get_current_applied_template

  current_deletion_policy=$(jq -r '.Resources | to_entries[] | select(.key | startswith("Cluster") and (. | startswith("ClusterSecretAttachment")| not) and (. | startswith("ClusterInstance") | not)) | .value.DeletionPolicy' current-cfn.json)
  new_deletion_policy=$(jq -r '.Resources | to_entries[] | select(.key | startswith("Cluster") and (. | startswith("ClusterSecretAttachment")| not) and (. | startswith("ClusterInstance") | not)) | .value.DeletionPolicy' cfn.json)

  if [ "${current_deletion_policy}" != "${new_deletion_policy}" ]; then
    echo "DeletionPolicy changed, updating stack..."
    aws cloudformation update-stack --stack-name "${STACK_NAME}" --template-body file://cfn.yaml --capabilities CAPABILITY_IAM --capabilities CAPABILITY_NAMED_IAM --no-cli-pager
    aws cloudformation wait stack-update-complete --stack-name "${STACK_NAME}" --no-cli-pager
    event_success
  fi
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
set -x
}

delete_stack() {
    cdk_synth
    if $(grep 'DeletionProtection: true' cfn.yaml > /dev/null 2>&1); then
        echo "DeletionProtection is enabled, update acorn app with '--deletion-protection=false' to delete this stack..."
        exit 1
    fi
    get_current_applied_template
    current_deletion_protection=$(jq -r '.Resources | to_entries[] | select(.key | startswith("Cluster") and (. | startswith("ClusterSecretAttachment")| not) and (. | startswith("ClusterInstance") | not)) | .value.Properties.DeletionProtection' current-cfn.json)

    # If current protection is true, and we are trying to set to false, run update
    if [ "${current_deletion_protection}" = "true" ]; then
      echo "Need to update stack to disable deletion protection..."
      aws cloudformation deploy --template-file cfn.yaml --stack-name "${STACK_NAME}" --capabilities CAPABILITY_IAM --capabilities CAPABILITY_NAMED_IAM --no-fail-on-empty-changeset --no-cli-pager
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
elif [ "$STATUS" = "ROLLBACK_FAILED" ]; then
  aws cloudformation rollback-stack --stack-name "${STACK_NAME}"
  aws cloudformation wait stack-rollback-complete --stack-name "${STACK_NAME}" --no-cli-pager --retain-except-on-create
fi

# Run CloudFormation
apply_and_render
if [ "${ACORN_EVENT}" = "update" ]; then
  update_stack_for_deletion_policy_only_updates
fi
exit 0