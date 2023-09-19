# AWS Elasticache Redis Service Acorn

Run an Elasticache Redis cluster as an Acorn with a single click or command.

## Usage

From the CLI you can run the following command to create an Elasticache Redis cluster.

```shell
acorn run -n elasticache-redis-cluster ghcr.io/acorn-io/elasticache/redis:v0.#.#
```

From an Acornfile you can create the cluster by using the acorn too.
```cue
services: redis: {
     image: "ghcr.io/acorn-io/elasticache/redis:v0.#.#"
}
containers: app: {
     build: context: "./"
     ports: publish: ["5000/http"]
     env: {
              REDIS_HOST: "@{service.redis.address}"
              REDIS_PORT: "@{service.redis.data.port}"
              REDIS_PASSWORD: "@{service.redis.secrets.admin.token}"
      }
}
```

To run from source can run `acorn run .` in this directory.

## Quirks

TLS is enabled for connections to the cluster. Some clients will fail to connect unless you specifically enable ssl/tls in their settings.

## Args

| Name               | Description                                                                                                                                                                   | Type   | Default         |
|--------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------|-----------------|
| clusterName        | Name to assign the Elasticache cluster during creation.                                                                                                                       | string | Redis           | 
| tags               | Key value pairs to apply to all resources.                                                                                                                                    | object | {}              | 
| deletionProtection | Prevents the cluster from being deleted when set to true.                                                                                                                     | bool   | false           | 
| nodeType           | The cache node type used in the elasticache cluster. See [elasticache pricing](https://aws.amazon.com/elasticache/pricing/) for a list of options.                            | string | cache.t4g.micro | 
| numNodes           | The number of cache nodes used in the elasticache cluster. Automatic failover is enabled for values >1. Cluster mode is disabled so it's a single primary with read replicas. | int    | 1               | 

## Output Services

```cue
services: admin: {
  default: true
  address: "${ADDRESS}"
  ports: [${PORT}]
  secrets: ["admin"]
  data: {
    clusterName: "${CLUSTER_NAME}"
    address: "${ADDRESS}"
    port: "${PORT}"
  }
}

secrets: "admin": {
  type: "token"
  data: {
    token: "${TOKEN}"
  }
}
```
