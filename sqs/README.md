# SQS Acorn

## Description

This Acorn provisions an AWS SQS queue.

## Usage

From the CLI you can deploy the default SQS queue using the following command:

```bash
acorn run ghcr.io/acorn-io/aws/sqs:v0.#.#
```

From the Acornfile you can launch an SQS queue using the following:

```cue
services: queue: {
    image: ghcr.io/acorn-io/aws/sqs:v0.#.#
}

containers: publisher: {
    image: "app"
    consumes: ["queue.publisher"]
    env: QUEUE_NAME: "@{services.queue.data.name}"
    env: QUEUE_URL:  "@{services.queue.data.url}"
}

container: subscriber: {
    image: "app"
    consumes: ["queue.subscriber"]
    env: QUEUE_NAME: "@{services.queue.data.name}"
    env: QUEUE_URL:  "@{services.queue.data.url}"
}

```

The above Acornfile shows two containers consuming different roles from the SQS Acorn. The publisher container has the ability to publish messages to the queue, while the subscriber container has the ability to receive messages from the queue.

There is a default service in the SQS Acorn in this case it is named `queue` and it is an alias for `queue.admin` which has administrative rights to the queue.

## Arguments

| Name | Description | Type |
|------|-------------|------|
| queueName | Name of queue. Defaults to acorn.externalID | string |
| fifo | Enable FIFO for the queue | bool |
| visibilityTimeout | Duration in seconds. Default is 30 seconds. | int |
| contentBasedDeduplication | Fifo Queue Option Only: ContentBasedDeduplication is a boolean that enables content-based deduplication. | bool |
| dataKeyReuse | Amount of time in seconds SQS reuses data key before calling KMS again | int |
| maxReceiveCount | Number of times a message can be unsuccessfully dequeued before being sent to the dead letter queue. A number >0 will create a new deadletter queue | int |
| encryptionMasterKey | KMS Key arn to use for encryption. Default is to use Amazon SQS key | string |
| tags | Key value pairs to apply to all AWS resources created by this Acorn | object |

## Service Outputs

The SQS Acorn provides three roles for use by applications. These can be used to provide least privilege access to your SQS queue from each container.

```cue
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
```
