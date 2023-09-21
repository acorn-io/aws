# AWS S3 (Simple Storage Service)

Create a S3 bucket as an Acorn with a single click or command.

## Usage

From the CLI you can run the following command to create a S3 bucket.

```shell
acorn run -n s3-bucket ghcr.io/acorn-io/aws/s3:v0.#.#
```

From an Acornfile you can create the bucket using the S3 acorn too.
```cue
services: s3: {
	image: "ghcr.io/acorn-io/aws/s3:v0.#.#"
}
containers: app: {
    build: context: "./"
    ports: publish: ["8080/http"]
    consumes: ["s3"]
    env: {
        BUCKET_URL: "@{service.s3.data.url}"
        BUCKET_NAME: "@{service.s3.data.name}"
        BUCKET_ARN: "@{service.s3.data.arn}"
    }
}
```

To run from source you can run `acorn run .` in this directory.

## Args

| Name               | Description                                                                                            | Type   | Default  |
|--------------------|--------------------------------------------------------------------------------------------------------|--------|----------|
| bucketName         | Name assigned to the bucket during creation.                                                           | string | MyBucket |
| versioned          | [Versioning](https://docs.aws.amazon.com/AmazonS3/latest/userguide/Versioning.html) is enabled if true | bool   | true     |
| tags               | Key value pairs to apply to all resources.                                                             | object | {}       |
| deletionProtection | Allows the bucket to be deleted when false.                                                            | bool   | false    |

## Output Services

```cue
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
```
