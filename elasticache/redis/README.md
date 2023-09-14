# AWS Elasticache Redis Service Acorn

Run an Elasticache Redis cluster as an Acorn with a single click or command.

## Quirks

TLS is enabled for connections to the cluster. Some clients will fail to connect unless you specifically enable ssl/tls in their settings.

## Args

```
      --cluster-name string                 Name to assign the elasticache cluster to creation.
      --tags string                         Key value pairs to apply to all resources.
      --deletion-protection bool            Prevents the cluster from being deleted when set to true. Default value is false.
      --node-type string                    The cache node type used in the elasticache cluster. Default value is "cache.t4g.micro". See https://aws.amazon.com/elasticache/pricing/ for a list of options.
      --num-nodes int                       The number of cache nodes use in the elasticache cluster. Default value is 1. Automatic failover is enabled for values >1. Cluster mode is disabled so it's a single primary with read replicas.
```

## Running from source 

`acorn run .` in this directory
