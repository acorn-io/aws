# AWS DynamoDB Service Acorn

Create a DynamoDB table as an Acorn with a single click or command.

## Usage

From the CLI you can run the following command to create the DynamoDB table.

```shell
acorn run -n elasticache-redis-cluster ghcr.io/acorn-io/aws/dynamodb:v0.#.#
```

From an Acornfile you can create the table by using the DynamoDB acorn too.
```cue
services: ddb: {
     image: "ghcr.io/acorn-io/aws/dynamodb:v0.#.#"
}
containers: app: {
     build: context: "./"
     ports: publish: ["8080/http"]
     consumes: ["ddb"]
     env: {
          TABLE_NAME: "@{service.ddb.data.name}"
          TABLE_ARN: "@{service.ddb.data.arn}"
     }
}
```

To run from source can run `acorn run .` in this directory.

## Args

| Name                 | Description                                                                                                                                                                                                                  | Type   | Default |
|----------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------|---------|
| tableName            | Name to assign the table during creation.                                                                                                                                                                                    | string |         | 
| partitionKey         | Key used to partition records.                                                                                                                                                                                               | string | id      | 
| partitionKeyType     | Type of the partition key. BINARY, STRING, and NUMBER are the valid values. See https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.NamingRulesDataTypes.html#HowItWorks.DataTypes for more details. | string | STRING  | 
| sortKey              | Key used to sort partitioned records.                                                                                                                                                                                        | string |         | 
| sortKeyType          | Type of the sort key. BINARY, STRING, and NUMBER are the valid values. See https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.NamingRulesDataTypes.html#HowItWorks.DataTypes for more details.      | string | STRING  | 
| tags                 | Key value pairs to apply to all resources.                                                                                                                                                                                   | object | {}      |
| deletionProtection   | Must be set to false to enable deletion of the table.                                                                                                                                                                        | bool   | false   |
| skipSnapshotOnDelete | Skip the final table snapshot before deletion if set to true.                                                                                                                                                                | bool   | false   |

## Output Services

```cue
services: {
  "admin": {
    consumer: permissions: rules: [{
      apiGroups: ["aws.acorn.io"]
      verbs: ["dynamodb:*"]
      resources: ["${arn}", "${arn}/*"]
    }]
    data: {
      arn: "${arn}"
      name: "${name}"
    }
  }

  "readonly": {
     name: "DynamoDB Reader"
     generated: job: "apply"
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
       arn: "${arn}"
       name: "${name}"
     }
  }

  "writeonly": {
     name: "DynamoDB Writer"
     generated: job: "apply"
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
       arn: "${arn}"
       name: "${name}"
     }
  }

  "readwrite": {
     name: "DynamoDB Reader Writer"
     generated: job: "apply"
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
       arn: "${arn}"
       name: "${name}"
     }
  }
}
```
