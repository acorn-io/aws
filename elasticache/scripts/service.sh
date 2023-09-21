#!/bin/bash

if [ ! -f outputs.json ]; then
   echo "outputs.json file not found!"
   exit 1
fi

# Render Output
CLUSTER_NAME=$(jq -r '.[] | select(.OutputKey=="clustername")|.OutputValue' outputs.json)
CLUSTER_ARN=$(jq -r '.[] | select(.OutputKey=="clusterarn")|.OutputValue' outputs.json)
ADDRESS=$(jq -r '.[] | select(.OutputKey=="address")|.OutputValue' outputs.json)
PORT=$(jq -r '.[]| select(.OutputKey=="port")|.OutputValue' outputs.json)
TOKEN_ARN=$(jq -r '.[]| select(.OutputKey=="tokenarn")|.OutputValue' outputs.json)

cat > /run/secrets/output <<EOF
services: admin: {
  default: true
  address: "${ADDRESS}"
  ports: [${PORT}]
EOF

# Only add secrets if TOKEN_ARN is not empty
if [ -n "${TOKEN_ARN}" ]; then
  TOKEN="$(aws --output text secretsmanager get-secret-value --secret-id "${TOKEN_ARN}" --query 'SecretString')"
  echo '  secrets: ["admin"]' >> /run/secrets/output
  cat >> /run/secrets/output <<EOF
  data: {
    clusterName: "${CLUSTER_NAME}"
    clusterArn: "${CLUSTER_ARN}"
    address: "${ADDRESS}"
    port: "${PORT}"
  }
}

secrets: "admin": {
  type: "token"
  data: {
    token: "${TOKEN}"
  }
}
EOF
else
  cat >> /run/secrets/output <<EOF
  data: {
    clusterName: "${CLUSTER_NAME}"
    clusterArn: "${CLUSTER_ARN}"
    address: "${ADDRESS}"
    port: "${PORT}"
  }
}
EOF
fi
