#!/bin/bash

if [ ! -f outputs.json ]; then
    echo "outputs.json file not found!"
    exit 1
fi

# Render Output
arn=$(jq -r '.[] | select(.OutputKey=="KMSKeyArn") | .OutputValue' outputs.json)

cat > /run/secrets/output<<EOF
services: "key": {
    default: true
    consumer: permissions: rules: [{
        apiGroups: ["aws.acorn.io"]
        verbs: [
            "kms:Decrypt",
            "kms:DescribeKey",
            "kms:Encrypt",
            "kms:GenerateDataKey",
            "kms:GenerateDataKeyPair",
            "kms:GenerateDataKeyPairWithoutPlaintext",
            "kms:GenerateMac",
            "kms:GenerateRandom",
            "kms:GetKeyPolicy",
            "kms:GetKeyRotationStatus",
            "kms:GetPublicKey",
            "kms:ListAliases",
            "kms:ListGrants",
            "kms:ListKeyPolicies",
            "kms:ListResourceTags",
            "kms:ListRetirableGrants",
            "kms:ReEncryptFrom",
            "kms:ReEncryptTo",
            "kms:Sign",
            "kms:Verify",
            "kms:VerifyMac",
        ]
        resources: ["${arn}"]
    }]
    data: arn: "${arn}"
}
EOF
