#!/bin/bash

if [ ! -f outputs.json ]; then
    echo "outputs.json file not found!"
    exit 1
fi

# Render Output
arn=$(jq -r '.[] | select(.OutputKey=="IAMRoleArn") | .OutputValue' outputs.json)

cat > /run/secrets/output<<EOF
services: "role": data: arn: "${arn}"
EOF
