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
  "s3-bucket-read": {
    address: "${address}"
    consumer: permissions: rules: [{
      apiGroups: ["aws.acorn.io"]
      verbs: ["s3:Get*", "s3:List*"]
      resources: ["${arn}"]
    }]
    data: {
      arn: "${arn}"
      arn: "${url}"
      proto: "${proto}"
      uri: "${uri}"
    }
  }

  "s3-bucket-write": {
    address: "${address}"
    consumer: permissions: rules: [{
      apiGroups: ["aws.acorn.io"]
      verbs: ["s3:Put*", "s3:AbortMultipartUpload"]
      resources: ["${arn}"]
    }]
    data: {
      arn: "${arn}"
      arn: "${url}"
      proto: "${proto}"
      uri: "${uri}"
    }
  }

  "s3-bucket-admin": {
    address: "${address}"
    consumer: permissions: rules: [{
      apiGroups: ["aws.acorn.io"]
      verbs: ["s3:*"]
      resources: ["${arn}"]
    }]
    data: {
      arn: "${arn}"
      arn: "${url}"
      proto: "${proto}"
      uri: "${uri}"
    }
  }
}
EOF
