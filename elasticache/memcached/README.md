# AWS Elasticache Memcached Service Acorn

Run an Elasticache Memcached cluster as an Acorn with a single click or command.

## Usage

From the CLI you can run the following command to create an Elasticache Memcached cluster.

```shell
acorn run -n memcached-cluster ghcr.io/acorn-io/elasticache/memcached:v0.#.#
```

From an Acornfile you can create the cluster by using the acorn too.
```cue
services: memcached: {
     image: "ghcr.io/acorn-io/elasticache/memcached:v0.#.#"
}
containers: app: {
     build: context: "./"
     ports: publish: ["8080/http"]
     env: {
              MEMCACHED_HOST: "@{service.memcached.address}"
              MEMCACHED_PORT: "@{service.memcached.data.port}"
     }
}
```

To run from source you can run `acorn run .` in this directory.

## Args

| Name               | Description                                                                                                                                        | Type   | Default         |
|--------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|--------|-----------------|
| clusterName        | Name to assign the Elasticache cluster during creation.                                                                                            | string | Memcached       |
| tags               | Key value pairs to apply to all resources.                                                                                                         | object | {}              |
| deletionProtection | Prevents the cluster from being deleted when set to true.                                                                                          | bool   | false           |
| nodeType           | The cache node type used in the elasticache cluster. See [elasticache pricing](https://aws.amazon.com/elasticache/pricing/) for a list of options. | string | cache.t4g.micro |
| numNodes           | The number of cache nodes used in the elasticache cluster.                                                                                         | int    | 1               |
| transitEncryption  | Enables TLS for connections to the cluster when true.                                                                                              | bool   | false           |

## Running from source 

`acorn run .` in this directory

## Output Services

```cue
services: admin: {
  default: true
  address: "${ADDRESS}"
  ports: [${PORT}]
  secrets: ["admin"]
  data: {
    clusterName: "${CLUSTER_NAME}"
    clusterArn: "${CLUSTER_ARN}"
    address: "${ADDRESS}"
    port: "${PORT}"
  }
}
```
