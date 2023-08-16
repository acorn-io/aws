#!/bin/bash

STACK_NAME="${ACORN_EXTERNAL_ID}"
# Check if outputs.json exists
if [ ! -f outputs.json ]; then
    echo "No outputs.json found. Exiting."
    exit 1
fi


# Render Output
url="$(jq -r '.[] | select(.OutputKey=="QueueURL")   |.OutputValue' outputs.json)"
arn="$(jq -r '.[] | select(.OutputKey=="QueueARN")|.OutputValue' outputs.json )"
name="$(jq -r '.[] | select(.OutputKey=="QueueName")|.OutputValue' outputs.json )"

proto="${url%%://*}"
no_proto="${url#*://}"
address="${no_proto%%/*}"
uri="${no_proto#*$address}"
record_success


cat > /run/secrets/output <<EOF
services: {
    admin: {
        default: true
        address: "${address}"
        consumer: permissions: rules: [{
            apiGroup: "aws.acorn.io"
		    verbs: [
			    "sqs:*",
		    ]
		    resources: ["${arn}"]
        }]
        data: {
            arn: "${arn}"
            proto: "${proto}"
            url: "${url}"
            uri: "${uri}"
            name: "${name}"
        }
    }
    publisher: {
        address: "${address}"
        consumer: permissions: rules: [{
            apiGroup: "aws.acorn.io"
            verbs: [
                "sqs:GetQueueUrl",
                "sqs:SendMessage",
            ]
            resources: ["${arn}"]
        }]
        data: {
            arn: "${arn}"
            proto: "${proto}"
            url: "${url}"
            uri: "${uri}"
            name: "${name}"
        }
    }
    subscriber: {
        address: "${address}"
        consumer: permissions: rules: [{
            apiGroup: "aws.acorn.io"
            verbs: [
                "sqs:ReceiveMessage",
                "sqs:DeleteMessage",
                "sqs:ChangeMessageVisibility",
                "sqs:GetQueueUrl",
            ]
            resources: ["${arn}"]
        }]
        data: {
            arn: "${arn}"
            proto: "${proto}"
            url: "${url}"
            uri: "${uri}"
            name: "${name}"
        }
    }
}
EOF
