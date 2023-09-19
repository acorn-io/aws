#!/bin/bash

# Render Output
PORT="$(          jq -r '.[] | select(.OutputKey=="port")            |.OutputValue' outputs.json )"
ADDRESS="$(       jq -r '.[] | select(.OutputKey=="host")            |.OutputValue' outputs.json )"
ADMIN_USERNAME="$(jq -r '.[] | select(.OutputKey=="adminusername")   |.OutputValue' outputs.json )"
PASSWORD_ARN="$(  jq -r '.[] | select(.OutputKey=="adminpasswordarn")|.OutputValue' outputs.json )"

ADMIN_PASSWORD="$(aws --output json secretsmanager get-secret-value --secret-id "${PASSWORD_ARN}" --query 'SecretString' | jq -r .|jq -r .password)"

cat > /run/secrets/output <<EOF
services: rds: {
  default: true
  address: "${ADDRESS}"
  ports: [${PORT}]
  data: {
    address: "${ADDRESS}"
    port: "${PORT}"
    dbName: "${DB_NAME}"
  }
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