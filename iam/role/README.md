# IAM Role Service Acorn

This Service Acorn creates a CloudFormation stack containing the given IAM Role and inline Policy.

## Limitations

Currently, this Service Acorn only supports trust relationships with a single ARN.
The ARN can refer to an AWS account, an IAM user, an IAM role, or specific assumed-role sessions.

## Usage

### Running the Acorn

```
acorn run ghcr.io/acorn-io/aws/iam/role:v0.1.0 \
  --roleName="my-role" \
  --policy @policy.json \
  --trustedArn="<arn>" \
  --externalIds="external-id-one,external-id-two" \
  --maxSessionDurationMinutes=120
```

### Using the service in an Acornfile

```cue
services: role: {
    image: "ghcr.io/acorn-io/aws/iam/role:v0.1.0"
    serviceArgs: {
        roleName:   "my-role"
        trustedArn: "<arn>"
        // This is an example policy:
        policy: {
            Version: "2012-10-17"
            Statement: [
                {
                    Effect: "Allow"
                    Action: [
                        "s3:ListBucket",
                    ]
                    Resource: [
                        "arn:aws:s3:::my-bucket",
                    ]
                },
                {
                    Effect: "Allow"
                    Action: [
                        "s3:GetObject",
                    ]
                    Resource: [
                        "arn:aws:s3:::my-bucket/*",
                    ]
                },
            ]
        }
    }
}

containers: mycontainer: {
    image: "<image>"
    env: ROLE_ARN: "@{services.role.data.arn}"
}
```

### Arguments

| Name                          | Description                                                                            | Required | Default                  |
|-------------------------------|----------------------------------------------------------------------------------------|----------|--------------------------|
| `--stackName`                 | The name of the CloudFormation stack to create.                                        | No       | (generated)              |
| `--roleName`                  | The name of the IAM role to create.                                                    | No       | (generated)              |
| `--policy`                    | The IAM policy to attach to the role as an inline policy. This must be in JSON format. | Yes      |                          |
| `--trustedArn`                | The ARN of the entity that can assume the role.                                        | Yes      |                          |
| `--externalIds`               | A comma-separated list of external IDs to use in the trust relationship.               | No       | (none)                   |
| `--maxSessionDurationMinutes` | The maximum session duration in minutes for the role.                                  | No       | 60                       |
| `--path`                      | The path in which to create the Role.                                                  | No       | "/"                      |
| `--tags`                      | Tags to attach to the Role.                                                            | No       | (none)                   |
| `--description`               | Description to attach to the Role.                                                     | No       | "Acorn created IAM Role" |

### Outputs

| Name  | Description                  |
|-------|------------------------------|
| `arn` | The ARN of the created role. |
