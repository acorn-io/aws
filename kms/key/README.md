# KMS Key Service Acorn

This Service Acorn creates a CloudFormation stack containing the given KMS Key.

## Limitations

Currently, this Service Acorn only supports adding a single ARN as an admin for the key.

## Usage

### Running the Acorn

```
acorn run ghcr.io/acorn-io/aws/kms/key:v0.1.0 \
  --key-name="my-key" \
  --key-alias="my-key" \
  --admin-arn="<arn>" \
  --description="Example key for encryption and decryption" \
  --key-spec="RSA_4098" \
  --key-usage="ENCRYPT_DECRYPT" \
  --pending-window-days=10 \
  --key-policy @policy.json
```

### Using the service in an Acornfile

```cue
services: key: {
    image: "ghcr.io/acorn-io/aws/kms/key:v0.1.0"
    serviceArgs: {
        keyName:           "my-key"
        keyAlias:          "my-key"
        adminArn:          "<arn>"
        description:       "Example key for encryption and decryption"
        keySpec:           "RSA_4098"
        keyUsage:          "ENCRYPT_DECRYPT"
        pendingWindowDays: 10
        tags: "my-tag": "my-tag-value"

        // This is an example policy:
        keyPolicy: {
            Version: "2012-10-07"
            Statement: [
                {
                    Effect: "Allow"
                    Principal: AWS: "arn:aws:iam::<account ID>:root"
                    Action:   "kms:*"
                    Resource: "*"
                },
            ]
        }
    }
}

containers: mycontainer: {
    image:    "<image>"
    consumes: ["key"]
    env: KEY_ARN: "@{services.key.data.arn}"
}

```

### Arguments

| Name                    | Description                                                                                                                                                             | Required | Default                                           |
|-------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|---------------------------------------------------|
| `--key-name`            | The name of the key in the CloudFormation stack.                                                                                                                        | No       | (generated)                                       |
| `--key-alias`           | The alias (friendly name) to give to the key.                                                                                                                           | No       | `@{acorn.name}-@{acorn.account}-@{acorn.project}` |
| `--admin-arn`           | The ARN of a user to set as the administrator of the key. You can specify AWS accounts, IAM users, Federated SAML users, IAM roles, and specific assumed-role sessions. | No       | (none)                                            |
| `--description`         | Description to attach to the key.                                                                                                                                       | No       | "Acorn created KMS Key"                           |
| `--key-spec`            | The type of key to create.                                                                                                                                              | Yes      | `SYMMETRIC_DEFAULT`                               |
| `--key-usage`           | The usage of the key. Each key spec only supports certain usages. See table below for details.                                                                          | Yes      | `ENCRYPT_DECRYPT`                                 |
| `--pending-window-days` | The time (in days) that must pass after key deletion is requested before the key is deleted. Must be between 7 and 30 (inclusive)                                       | Yes      | 7                                                 |
| `--key-policy`          | The key policy to attach to the key. This must be in JSON format.                                                                                                       | No       | (created by AWS)                                  |
| `--tags`                | Tags to attach to the key.                                                                                                                                              | No       | (none)                                            |

#### Key Specs and Usages

| Key Spec            | Supported Key Usages             |
|---------------------|----------------------------------|
| `SYMMETRIC_DEFAULT` | `ENCRYPT_DECRYPT`                |
| `RSA_2048`          | `ENCRYPT_DECRYPT`, `SIGN_VERIFY` |
| `RSA_3072`          | `ENCRYPT_DECRYPT`, `SIGN_VERIFY` |
| `RSA_4096`          | `ENCRYPT_DECRYPT`, `SIGN_VERIFY` |
| `ECC_NIST_P256`     | `SIGN_VERIFY`                    |
| `ECC_NIST_P384`     | `SIGN_VERIFY`                    |
| `ECC_NIST_P521`     | `SIGN_VERIFY`                    |
| `ECC_SECG_P256K1`   | `SIGN_VERIFY`                    |
| `HMAC_224`          | `GENERATE_VERIFY_MAC`            |
| `HMAC_256`          | `GENERATE_VERIFY_MAC`            |
| `HMAC_384`          | `GENERATE_VERIFY_MAC`            |
| `HMAC_512`          | `GENERATE_VERIFY_MAC`            |

Source: https://pkg.go.dev/github.com/aws/aws-cdk-go/awscdk/v2/awskms@v2.96.0#KeySpec

### Outputs

| Name  | Description                 |
|-------|-----------------------------|
| `arn` | The ARN of the created key. |
