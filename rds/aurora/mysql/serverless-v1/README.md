# Aurora MySQL Serverless V1

This Acorn creates an Aurora MySQL serverless cluster running on AWS RDS service. AWS is moving towards the Serverless V2 solution, which you can find [here](https://github.com/acorn-io/aws/tree/main/rds/aurora/mysql/serverless-v2). This Acorn is best used when you have a variable workload that sometimes is completely unused. This Acorn may not launch in all regions or with all configurations.

If you have a stable workload, you should evaluate the cost of serverless vs. cluster based Aurora MySQL.

## Usage

From the CLI you can run the following command to create an Aurora MySQL cluster:

```shell
acorn run -n rds-mysql-cluster ghcr.io/acorn-io/aws/rds/aurora/mysql/serverless-v1:v1.#.#
```

From an Acornfile you can use the following Acorn:

```cue
services: "rds-mysql-cluster": {
    image: ghcr.io/acorn-io/aws/rds/aurora/mysql/serverless-v1:v1.#.#
}

containers: wp: {
    image: wordpress:latest
    ports: publish: "80/http"
    env: {
        WORDPRESS_DB_HOST: "@{services.rds-mysql-cluster.address}"
        WORDPRESS_DB_USER: "@{services.rds-mysql-cluster.secrets.amdin.username}"
        WORDPRESS_DB_PASSWORD: "@{services.rds-mysql-cluster.secrets.admin.password}"
        WORDPRESS_DB_NAME: "instance"
    }
}
```

## Arguments

| Name | Description | Type |
|------|-------------|------|
| adminUsername | Name of the root/admin user. Default is admin. | string |
| username | Name of an additional user to create. This user will have complete access to the database
If left blank, no additional user will be created. | string |
| dbName | Name of the database instance. Default is instance. | string |
| deletionProtection | Deletion protection, you must set to false in order for the RDS db to be deleted. Default is false. | bool |
| parameters | RDS MySQL Database Parameters to apply to the cluster. Must be k/v string pairs(ex. max_connections: "1000"). | object |
| auroraCapacityUnitsMin | Aurora Capacity Units minimum value must be 1, 2, 4, 8, 16, 32, 64, 128, 256, 384. Default is 4. | int |
| auroraCapacityUnitsMax | Aurora Capacity Units maximum value must be 1, 2, 4, 8, 16, 32, 64, 128, 256, 384. Default is 8. | int |
| autoPauseDurationMinutes | Time in minutes to pause Aurora serverless-v1 DB cluster after it's been idle. Default is 10 set to 0 to disable. | int |
| skipSnapshotOnDelete | Do not take a final snapshot on delete or update and replace operations. Default is false. If skip is enabled the DB will be gone forever if deleted or replaced. | bool |
| tags | Key value pairs of tags to apply to the RDS cluster and all other resources. | object |

## Output Services

```cue
services: rds: {
  default: true
  address: "${ADDRESS}"
  ports: [${PORT}]
  secrets: ["admin"]
  data: {
    address: "${ADDRESS}"
    port: "${PORT}"
    dbName: "${DB_NAME}"
  }
}

secrets: "admin": {
 type: "basic"
 data: {
    username: "${ADMIN_USERNAME}"
    password: "${ADMIN_PASSWORD}"
 }
}

// If username is set, create a user secret
secrets: user: {
    type: "basic"
    data: {
        username: "${USERNAME}"
        password: "${PASSWORD}"
    }
}
```
