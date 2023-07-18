# AMP

## Args

```

Volumes:   <none>
Secrets:   <none>
Containers: <none>
Ports:     <none>

      --profile string          
      --tags string             Key value pairs to tag the workspace.
      --workspace-name string   Name of workspace to create, only use if you need a specifc name.
                                Otherwise let Acorn generate one.

```

## Service Output

```
{
  "amp": {
    "default": true,
    "address": "${address}",
    "data": {
      "arn": "${arn}",
      "proto": "${proto}",
      "uri": "${uri}",
      "url": "${url}"
    }
  }
}

```

## permissions

```
{
  "apply": {
    "rules": [
      {
        "verbs": [
          "cloudformation:DescribeStacks",
          "cloudformation:CreateChangeSet",
          "cloudformation:DescribeChangeSet",
          "cloudformation:DescribeStackEvents",
          "cloudformation:ExecuteChangeSet",
          "cloudformation:PreviewStackUpdate",
          "cloudformation:UpdateStack",
          "cloudformation:GetTemplateSummary",
          "cloudformation:DeleteStack",
          "aps:*"
        ],
        "apiGroups": [
          "aws.acorn.io"
        ],
        "resources": [
          "*"
        ]
      }
    ]
  }
}

```
