#!/bin/bash

if [ ! -f outputs.json ]; then
   echo "outputs.json file not found!"
   exit 1
fi

# Render Output
name=$(jq -r '.[] | select(.OutputKey=="TableName")|.OutputValue' outputs.json)
arn=$(jq -r '.[] | select(.OutputKey=="TableARN")|.OutputValue' outputs.json)

cat > /run/secrets/output<<EOF
services: {
	admin: {
		consumer: permissions: rules: [{
			apiGroups: ["aws.acorn.io"]
			verbs: ["dynamodb:*"]
			resources: ["${arn}", "${arn}/*"]
		}]
		data: {
			arn:  "${arn}"
			name: "${name}"
		}
	}
	readonly: {
		name: "DynamoDB Reader"
		consumer: permissions: rules: [{
			apiGroups: ["aws.acorn.io"]
			verbs: [
				"dynamodb:BatchGetItem",
				"dynamodb:ConditionCheckItem",
				"dynamodb:GetItem",
				"dynamodb:GetRecords",
				"dynamodb:GetShardIterator",
				"dynamodb:PartiQLSelect",
				"dynamodb:Query",
				"dynamodb:Scan",
			]
			resources: ["${arn}", "${arn}/*"]
		}]
		data: {
			arn:  "${arn}"
			name: "${name}"
		}
	}
	writeonly: {
		name: "DynamoDB Writer"
		consumer: permissions: rules: [{
			apiGroups: ["aws.acorn.io"]
			verbs: [
				"dynamodb:BatchWriteItem",
				"dynamodb:PartiQLInsert",
				"dynamodb:PartiQLUpdate",
				"dynamodb:PutItem",
				"dynamodb:UpdateItem",
			]
			resources: ["${arn}", "${arn}/*"]
		}]
		data: {
			arn:  "${arn}"
			name: "${name}"
		}
	}
	readwrite: {
		name:    "DynamoDB Reader Writer"
		default: true
		consumer: permissions: rules: [{
			apiGroups: ["aws.acorn.io"]
			verbs: [
				"dynamodb:BatchWriteItem",
				"dynamodb:PartiQLInsert",
				"dynamodb:PartiQLUpdate",
				"dynamodb:PartiQLSelect",
				"dynamodb:PartiQLDelete",
				"dynamodb:PutItem",
				"dynamodb:UpdateItem",
				"dynamodb:DeleteItem",
				"dynamodb:BatchGetItem",
				"dynamodb:ConditionCheckItem",
				"dynamodb:GetItem",
				"dynamodb:GetRecords",
				"dynamodb:GetShardIterator",
				"dynamodb:Query",
				"dynamodb:Scan",
			]
			resources: ["${arn}", "${arn}/*"]
		}]
		data: {
			arn:  "${arn}"
			name: "${name}"
		}
	}
}
EOF
