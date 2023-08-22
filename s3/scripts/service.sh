#!/bin/bash

if [ ! -f outputs.json ]; then
   echo "outputs.json file not found!"
   exit 1
fi

# Render Output
url=$(jq -r '.[] | select(.OutputKey=="BucketURL")|.OutputValue' outputs.json)
arn=$(jq -r '.[]| select(.OutputKey=="BucketARN")|.OutputValue' outputs.json)
name=$(jq -r '.[]| select(.OutputKey=="BucketName")|.OutputValue' outputs.json)
proto="${url%%://*}"
no_proto="${url#*://}"
address="${no_proto%%/*}"
uri="${no_proto#*$address}"

cat > /run/secrets/output<<EOF
services: {
  "readwrite": {
    address: "${address}"
    consumer: permissions: rules: [{
      apiGroups: ["aws.acorn.io"]
      verbs: ["s3:Get*", "s3:List*", "s3:Put*", "s3:AbortMultipartUpload"]
      resources: ["${arn}", "${arn}/*"]
    }, {
      apiGroups: ["aws.acorn.io"]
      verbs: ["s3:ListBuckets"]
      resources: ["*"]
    }]
    data: {
      name: "${name}"
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
      verbs: ["s3:Get*", "s3:List*"]
      resources: ["${arn}", "${arn}/*"]
    }]
    data: {
      name: "${name}"
      arn: "${arn}"
      url: "${url}"
      proto: "${proto}"
      uri: "${uri}"
    }
  }

  "writeonly": {
    address: "${address}"
    consumer: permissions: rules: [{
      apiGroups: ["aws.acorn.io"]
      verbs: ["s3:Put*", "s3:AbortMultipartUpload"]
      resources: ["${arn}", "${arn}/*"]
    }]
    data: {
      name: "${name}"
      arn: "${arn}"
      url: "${url}"
      proto: "${proto}"
      uri: "${uri}"
    }
  }

  "admin": {
    address: "${address}"
    consumer: permissions: rules: [{
      apiGroups: ["aws.acorn.io"]
      verbs: ["s3:*"]
      resources: ["${arn}", "${arn}/*"]
    }, {
      apiGroups: ["aws.acorn.io"]
      verbs: ["s3:ListBuckets"]
      resources: ["*"]
    }]
    data: {
      name: "${name}"
      arn: "${arn}"
      url: "${url}"
      proto: "${proto}"
      uri: "${uri}"
    }
  }
}
EOF
