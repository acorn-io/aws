#!/bin/bash
set -e -o pipefail

arn="${1}"

echo Getting secret value for "$arn"
data="$(aws --region "${REGION:-$AWS_REGION}" --output json secretsmanager get-secret-value --secret-id "${arn}" | jq '{value: .SecretString}')"

if [ "$FORMAT" = "json" ]; then
  data="$(echo "$data" | jq -r .value)"
fi

cat > /run/secrets/output <<EOF
services: "secret-manager": {
    default: true
    secrets: ["secret-value"]
}

secrets: "item": {
    type: "opaque"
    data: ${data}
}
EOF
