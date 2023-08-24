#!/bin/bash

if [ ! -f outputs.json ]; then
    echo "outputs.json file not found!"
    exit 1
fi

# Render Output
url=$(jq -r '.[] | select(.OutputKey=="AMPEndpointURL")|.OutputValue' outputs.json)
arn=$(jq -r '.[]| select(.OutputKey=="AMPWorkspaceArn")|.OutputValue' outputs.json)
proto="${url%%://*}"
no_proto="${url#*://}"
address="${no_proto%%/*}"
uri="${no_proto#*$address}"

cat > /run/secrets/output<<EOF
services: {
    "admin": {
        default: true
        address: "${address}"
        consumer: permissions: rules: [{
           apiGroups: ["aws.acorn.io"]
           verbs: ["aps:*"] 
           resources: ["${arn}"]
        }]
        data: {
            arn: "${arn}"
            url: "${url}"
            proto: "${proto}"
            uri: "${uri}"
        }
    }
    "readonly": {
        address: "${address}"
        consumer: permissions: rules: [{
           apiGroups: ["aws.acorn.io"]
           verbs: [
            "aps:GetLabels",
            "aps:GetMetricMetadata",
            "aps:GetSeries",
            "aps:QueryMetrics",
           ]
           resources: ["${arn}"]
        }]
        data: {
            arn: "${arn}"
            url: "${url}"
            proto: "${proto}"
            uri: "${uri}"
        }
    }
    "remote-write": {
        address: "${address}"
        consumer: permissions: rules: [{
           apiGroups: ["aws.acorn.io"]
           verbs: [
            "aps:RemoteWrite",
           ]
           resources: ["${arn}"]
        }]
        data: {
            arn: "${arn}"
            url: "${url}"
            proto: "${proto}"
            uri: "${uri}"
        }
    }
}
EOF
